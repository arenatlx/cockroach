// Copyright 2013 Google Inc. All Rights Reserved.
// Copyright 2017 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// This code originated in the github.com/golang/glog package.

package log

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cli/exit"
	"github.com/cockroachdb/cockroach/pkg/util/log/logpb"
	"github.com/cockroachdb/cockroach/pkg/util/log/severity"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/logtags"
)

// logging is the global state of the logging setup.
var logging loggingT

// loggingT collects all the global state of the logging setup.
//
// TODO(knz): better separate global state and per-logger state.
type loggingT struct {
	config

	// allocation pool for entry formatting buffers.
	bufPool sync.Pool

	// interceptor is the configured InterceptorFn callback, if any.
	interceptor atomic.Value

	// vmoduleConfig maintains the configuration for the log.V and vmodule
	// facilities.
	vmoduleConfig vmoduleConfig

	// mu protects the remaining elements of this structure and is
	// used to synchronize logging.
	// mu should be held only for short periods of time and
	// its critical sections cannot contain logging calls themselves.
	mu struct {
		syncutil.Mutex

		// disableDaemons can be used to turn off both the GC and flush deamons.
		disableDaemons bool

		// exitOverride is used when shutting down logging.
		exitOverride struct {
			f         func(exit.Code, error) // overrides exit.WithCode when non-nil; testing only
			hideStack bool                   // hides stack trace; only in effect when f is not nil
		}

		// the Cluster ID is reported on every new log file so as to ease the correlation
		// of panic reports with self-reported log files.
		clusterID string

		// fatalCh is closed on fatal errors.
		fatalCh chan struct{}

		// active indicates that at least one event has been logged
		// to this logger already.
		active        bool
		firstUseStack string
	}
}

func init() {
	logging.bufPool.New = newBuffer
	logging.mu.fatalCh = make(chan struct{})
}

// loggerT represents the logging source for a given log channel.
type loggerT struct {
	// fileSink is the file sink for this channel.
	//
	// TODO(knz): Although the logging logic below is a-OK
	// with fileSink == nil (which means "no logging to files"),
	// the test code in this package still requires fileSink
	// to always be defined. This should be cleaned up to
	// clarify the semantics.
	fileSink *fileSink

	// whether or not to include redaction markers.
	// This is atomic because tests using TestLogScope might
	// override this asynchronously with log calls.
	redactableLogs syncutil.AtomicBool

	// syncWrites if true calls file.Flush and file.Sync on every log
	// write. This can be set per-logger e.g. for audit logging.
	//
	// Note that synchronization for all log files simultaneously can
	// also be configured via logging.syncWrites, see SetSync().
	syncWrites bool

	// logCounter supports the generation of a per-entry log entry
	// counter. This is needed in audit logs to hinder malicious
	// repudiation of log events by manually erasing log files or log
	// entries.
	logCounter EntryCounter

	// outputMu is used to coordinate output to the sinks, to guarantee
	// that the ordering of events the the same on all sinks.
	outputMu syncutil.Mutex
}

// EntryCounter supports the generation of a per-entry log entry
// counter. This is needed in audit logs to hinder malicious
// repudiation of log events by manually erasing log files or log
// entries.
type EntryCounter struct {
	// EnableMsgCount, if true, enables the production of entry
	// counters.
	EnableMsgCount bool
	// msgCount is the current value of the counter.
	msgCount uint64
}

// FatalChan is closed when Fatal is called. This can be used to make
// the process stop handling requests while the final log messages and
// crash report are being written.
func FatalChan() <-chan struct{} {
	return logging.mu.fatalCh
}

// s ignalFatalCh signals the listeners of l.mu.fatalCh by closing the
// channel.
// l.mu is not held.
func (l *loggingT) signalFatalCh() {
	l.mu.Lock()
	defer l.mu.Unlock()
	// Close l.fatalCh if it is not already closed (note that we're
	// holding l.mu to guard against concurrent closes).
	select {
	case <-l.mu.fatalCh:
	default:
		close(l.mu.fatalCh)
	}
}

// SetClusterID stores the Cluster ID for further reference.
//
// TODO(knz): This should not be configured per-logger.
// See: https://github.com/cockroachdb/cockroach/issues/40983
func SetClusterID(clusterID string) {
	// Ensure that the clusterID is logged with the same format as for
	// new log files, even on the first log file. This ensures that grep
	// will always find it.
	ctx := logtags.AddTag(context.Background(), "config", nil)
	addStructured(ctx, severity.INFO, 1, "clusterID: %s", []interface{}{clusterID})

	// Perform the change proper.
	logging.mu.Lock()
	defer logging.mu.Unlock()

	if logging.mu.clusterID != "" {
		panic("clusterID already set")
	}

	logging.mu.clusterID = clusterID
}

// outputLogEntry marshals a log entry proto into bytes, and writes
// the data to the log files. If a trace location is set, stack traces
// are added to the entry before marshaling.
func (l *loggerT) outputLogEntry(entry logpb.Entry) {
	if f, ok := logging.interceptor.Load().(InterceptorFn); ok && f != nil {
		f(entry)
		return
	}

	// Mark the logger as active, so that further configuration changes
	// are disabled. See IsActive() and its callers for details.
	setActive()
	var stacks []byte
	var fatalTrigger chan struct{}
	fileSink := l.getFileSink()

	if entry.Severity == severity.FATAL {
		logging.signalFatalCh()

		switch traceback {
		case tracebackSingle:
			stacks = getStacks(false)
		case tracebackAll:
			stacks = getStacks(true)
		}

		// Since the Fatal output will be copied to stderr below, it may
		// show up to a (human) observer through a different channel than
		// a file in the log directory. So remind them where to look for
		// more.
		if fileSink != nil {
			fileSink.mu.Lock()
			stacks = append(stacks, []byte(fmt.Sprintf(
				"\nFor more context, check log files in: %s\n", fileSink.mu.logDir))...)
			fileSink.mu.Unlock()
		}

		// Explain to the (human) user that we would like to hear from them.
		stacks = append(stacks, []byte(fatalErrorPostamble)...)

		// We don't want to hang forever writing our final log message. If
		// things are broken (for example, if the disk fills up and there
		// are cascading errors and our process manager has stopped
		// reading from its side of a stderr pipe), it's more important to
		// let the process exit than limp along.
		//
		// Note that we do not use os.File.SetWriteDeadline because not
		// all files support this (for example, plain files on a network
		// file system do not support deadlines but can block
		// indefinitely).
		//
		// https://github.com/cockroachdb/cockroach/issues/23119
		fatalTrigger = make(chan struct{})
		exitFunc := func(x exit.Code, _ error) { exit.WithCode(x) }
		logging.mu.Lock()
		if logging.mu.exitOverride.f != nil {
			if logging.mu.exitOverride.hideStack {
				stacks = []byte("stack trace omitted via SetExitFunc()\n")
			}
			exitFunc = logging.mu.exitOverride.f
		}
		logging.mu.Unlock()
		exitCalled := make(chan struct{})

		// This defer prevents outputLogEntry() from returning until the
		// exit function has been called.
		defer func() {
			<-exitCalled
		}()
		go func() {
			select {
			case <-time.After(10 * time.Second):
			case <-fatalTrigger:
			}
			exitFunc(exit.FatalError(), nil)
			close(exitCalled)
		}()
	}

	// The following buffers contain the formatted entry before it enters the sink.
	// We need different buffers because the different sinks use different formats.
	// For example, the fluent sink needs JSON, and the file sink does not use
	// the terminal escape codes that the stderr sink uses.
	var stderrBuf, fileBuf *buffer
	defer func() {
		// Release the buffers to the allocation pool upon returning from
		// this function.
		putBuffer(stderrBuf)
		putBuffer(fileBuf)
	}()

	// The following code constructs / populates the formatted entries
	// for each sink.
	// We only do the work if the sink is active and the filtering does
	// not eliminate the event.

	if entry.Severity >= logging.stderrThreshold.Get() {
		stderrBuf = logging.processForStderr(entry, stacks)
	}

	if fileSink != nil && entry.Severity >= fileSink.fileThreshold {
		fileBuf = logging.processForFile(entry, stacks)
	}

	// If any of the sinks is active, it is now time to send it out.

	if stderrBuf != nil || fileBuf != nil {
		// The critical section here exists so that the output
		// side effects from the same event (above) are emitted
		// atomically. This ensures that the order of logging
		// events is preserved across all sinks.
		l.outputMu.Lock()
		defer l.outputMu.Unlock()

		if stderrBuf != nil {
			if err := l.outputToStderr(stderrBuf.Bytes()); err != nil {
				// The external stderr log is unavailable.  However, stderr was
				// chosen by the stderrThreshold configuration, so abandoning
				// the stderr write would be a contract violation.
				//
				// We definitely do not like to lose log entries, so we stop
				// here. Note that exitLocked() shouts the error to both stderr
				// and the log file, so even though stderr is not available any
				// more, we'll keep a trace of the error in the file.
				l.exitLocked(err, exit.LoggingStderrUnavailable())
				return // unreachable except in tests
			}
		}

		if fileBuf != nil && fileSink.enabled.Get() {
			// NB: we need to check filesink.enabled a second time here in
			// case a test Scope() has disabled it asynchronously while we
			// were not holding outputMu above.
			if err := fileSink.output(
				l.syncWrites,                     /* doSync */
				entry.Severity == severity.FATAL, /* doFlush*/
				fileBuf.Bytes()); err != nil {
				// We definitely do not like to lose log entries, so we stop
				// here. Note that exitLocked() shouts the error to both stderr
				// and the log file, so even though the file is not available
				// any more, we'll likely keep a trace of the error in stderr.
				l.exitLocked(err, exit.LoggingFileUnavailable())
				return // unreachable except in tests
			}
		}
	}

	// Flush and exit on fatal logging.
	if entry.Severity == severity.FATAL {
		close(fatalTrigger)
		// Note: although it seems like the function is allowed to return
		// below when s == severity.FATAL, this is not so, because the
		// anonymous function func() { <-exitCalled } is deferred
		// above. That function ensures that outputLogEntry() will wait
		// until the exit function has been called. If the exit function
		// is exit.WithCode, it will never return, outputLogEntry()'s defer will
		// never complete and all is well. If the exit function was
		// overridden, then the client that has overridden the exit
		// function is expecting log.Fatal to return and all is well too.
	}
}

// getFileSink returns the fileSink if the file logger is configured
// to a valid directory name.
func (l *loggerT) getFileSink() *fileSink {
	if l.fileSink != nil && l.fileSink.enabled.Get() {
		return l.fileSink
	}
	return nil
}

// DumpStacks produces a dump of the stack traces in the logging output.
func DumpStacks(ctx context.Context) {
	allStacks := getStacks(true)
	// TODO(knz): This should really be a "debug" level, not "info".
	Infof(ctx, "stack traces:\n%s", allStacks)
}

func setActive() {
	logging.mu.Lock()
	defer logging.mu.Unlock()
	if !logging.mu.active {
		logging.mu.active = true
		logging.mu.firstUseStack = string(debug.Stack())
	}
}

// outputToStderr writes the provided entry and potential stack
// trace(s) to the process' external stderr stream.
func (l *loggerT) outputToStderr(b []byte) error {
	_, err := OrigStderr.Write(b)
	return err
}

const fatalErrorPostamble = `

****************************************************************************

This node experienced a fatal error (printed above), and as a result the
process is terminating.

Fatal errors can occur due to faulty hardware (disks, memory, clocks) or a
problem in CockroachDB. With your help, the support team at Cockroach Labs
will try to determine the root cause, recommend next steps, and we can
improve CockroachDB based on your report.

Please submit a crash report by following the instructions here:

    https://github.com/cockroachdb/cockroach/issues/new/choose

If you would rather not post publicly, please contact us directly at:

    support@cockroachlabs.com

The Cockroach Labs team appreciates your feedback.
`

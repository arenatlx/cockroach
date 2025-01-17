// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package sql

import (
	"testing"

	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/multiregion"
	"github.com/cockroachdb/cockroach/pkg/util/leaktest"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"
)

func TestZoneConfigForMultiRegionDatabase(t *testing.T) {
	defer leaktest.AfterTest(t)()

	testCases := []struct {
		desc         string
		regionConfig multiregion.RegionConfig
		expected     zonepb.ZoneConfig
	}{
		{
			desc: "one region, zone survival",
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_a",
			}, "region_a", descpb.SurvivalGoal_ZONE_FAILURE, descpb.InvalidID, descpb.DataPlacement_DEFAULT, nil),
			expected: zonepb.ZoneConfig{
				NumReplicas: proto.Int32(3),
				NumVoters:   proto.Int32(3),
				LeasePreferences: []zonepb.LeasePreference{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
				},
				Constraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
				},
				NullVoterConstraintsIsEmpty: true,
				VoterConstraints: []zonepb.ConstraintsConjunction{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
				},
			},
		},
		{
			desc: "two regions, zone survival",
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_a",
			}, "region_a", descpb.SurvivalGoal_ZONE_FAILURE, descpb.InvalidID, descpb.DataPlacement_DEFAULT, nil),
			expected: zonepb.ZoneConfig{
				NumReplicas: proto.Int32(4),
				NumVoters:   proto.Int32(3),
				LeasePreferences: []zonepb.LeasePreference{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
				},
				Constraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
				},
				NullVoterConstraintsIsEmpty: true,
				VoterConstraints: []zonepb.ConstraintsConjunction{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
				},
			},
		},
		{
			desc: "three regions, zone survival",
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
			}, "region_b", descpb.SurvivalGoal_ZONE_FAILURE, descpb.InvalidID, descpb.DataPlacement_DEFAULT, nil),
			expected: zonepb.ZoneConfig{
				NumReplicas: proto.Int32(5),
				NumVoters:   proto.Int32(3),
				LeasePreferences: []zonepb.LeasePreference{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
				Constraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_c"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
				},
				NullVoterConstraintsIsEmpty: true,
				VoterConstraints: []zonepb.ConstraintsConjunction{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
			},
		},
		{
			desc: "three regions, region survival",
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
			}, "region_b", descpb.SurvivalGoal_REGION_FAILURE, descpb.InvalidID, descpb.DataPlacement_DEFAULT, nil),
			expected: zonepb.ZoneConfig{
				NumReplicas: proto.Int32(5),
				NumVoters:   proto.Int32(5),
				LeasePreferences: []zonepb.LeasePreference{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"}},
					},
				},
				Constraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_c"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
				},
				NullVoterConstraintsIsEmpty: true,
				VoterConstraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 2,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
			},
		},
		{
			desc: "four regions, zone survival",
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
				"region_d",
			}, "region_b", descpb.SurvivalGoal_ZONE_FAILURE, descpb.InvalidID, descpb.DataPlacement_DEFAULT, nil),
			expected: zonepb.ZoneConfig{
				NumReplicas: proto.Int32(6),
				NumVoters:   proto.Int32(3),
				LeasePreferences: []zonepb.LeasePreference{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
				Constraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_c"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_d"},
						},
					},
				},
				NullVoterConstraintsIsEmpty: true,
				VoterConstraints: []zonepb.ConstraintsConjunction{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
			},
		},
		{
			desc: "four regions, region survival",
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
				"region_d",
			}, "region_b", descpb.SurvivalGoal_REGION_FAILURE, descpb.InvalidID, descpb.DataPlacement_DEFAULT, nil),
			expected: zonepb.ZoneConfig{
				NumReplicas: proto.Int32(5),
				NumVoters:   proto.Int32(5),
				LeasePreferences: []zonepb.LeasePreference{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
				Constraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_c"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_d"},
						},
					},
				},
				NullVoterConstraintsIsEmpty: true,
				VoterConstraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 2,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
			},
		},
		{
			desc: "one region, restricted placement",
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_a",
			}, "region_a", descpb.SurvivalGoal_ZONE_FAILURE, descpb.InvalidID, descpb.DataPlacement_RESTRICTED, nil),
			expected: zonepb.ZoneConfig{
				NumReplicas: proto.Int32(3),
				NumVoters:   proto.Int32(3),
				LeasePreferences: []zonepb.LeasePreference{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
				},
				Constraints:                 nil,
				NullVoterConstraintsIsEmpty: true,
				VoterConstraints: []zonepb.ConstraintsConjunction{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
				},
			},
		},
		{
			desc: "four regions, restricted placement",
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_a",
				"region_b",
				"region_c",
				"region_d",
			}, "region_a", descpb.SurvivalGoal_ZONE_FAILURE, descpb.InvalidID, descpb.DataPlacement_RESTRICTED, nil),
			expected: zonepb.ZoneConfig{
				NumReplicas: proto.Int32(3),
				NumVoters:   proto.Int32(3),
				LeasePreferences: []zonepb.LeasePreference{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
				},
				Constraints:                 nil,
				NullVoterConstraintsIsEmpty: true,
				VoterConstraints: []zonepb.ConstraintsConjunction{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			res, err := zoneConfigForMultiRegionDatabase(tc.regionConfig)
			require.NoError(t, err)
			require.Equal(t, tc.expected, res)
		})
	}
}

func protoRegionName(region catpb.RegionName) *catpb.RegionName {
	return &region
}

func TestZoneConfigForMultiRegionTable(t *testing.T) {
	defer leaktest.AfterTest(t)()

	testCases := []struct {
		desc           string
		localityConfig catpb.LocalityConfig
		regionConfig   multiregion.RegionConfig
		expected       zonepb.ZoneConfig
	}{
		{
			desc: "4-region global table with zone survival",
			localityConfig: catpb.LocalityConfig{
				Locality: &catpb.LocalityConfig_Global_{
					Global: &catpb.LocalityConfig_Global{},
				},
			},
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
				"region_d",
			}, "region_b", descpb.SurvivalGoal_ZONE_FAILURE, descpb.InvalidID, descpb.DataPlacement_DEFAULT, nil),
			expected: zonepb.ZoneConfig{
				GlobalReads:               proto.Bool(true),
				InheritedConstraints:      true,
				InheritedLeasePreferences: true,
			},
		},
		{
			desc: "4-region global table with region survival",
			localityConfig: catpb.LocalityConfig{
				Locality: &catpb.LocalityConfig_Global_{
					Global: &catpb.LocalityConfig_Global{},
				},
			},
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
				"region_d",
			}, "region_b", descpb.SurvivalGoal_REGION_FAILURE, descpb.InvalidID, descpb.DataPlacement_DEFAULT, nil),
			expected: zonepb.ZoneConfig{
				GlobalReads:               proto.Bool(true),
				InheritedConstraints:      true,
				InheritedLeasePreferences: true,
			},
		},
		{
			desc: "4-region regional by row table with zone survival",
			localityConfig: catpb.LocalityConfig{
				Locality: &catpb.LocalityConfig_RegionalByRow_{
					RegionalByRow: &catpb.LocalityConfig_RegionalByRow{},
				},
			},
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
				"region_d",
			}, "region_b", descpb.SurvivalGoal_ZONE_FAILURE, descpb.InvalidID, descpb.DataPlacement_DEFAULT, nil),
			expected: *(zonepb.NewZoneConfig()),
		},
		{
			desc: "4-region regional by row table with region survival",
			localityConfig: catpb.LocalityConfig{
				Locality: &catpb.LocalityConfig_RegionalByRow_{
					RegionalByRow: &catpb.LocalityConfig_RegionalByRow{},
				},
			},
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
				"region_d",
			}, "region_b", descpb.SurvivalGoal_ZONE_FAILURE, descpb.InvalidID, descpb.DataPlacement_DEFAULT, nil),
			expected: *(zonepb.NewZoneConfig()),
		},
		{
			desc: "4-region regional by table with zone survival on primary region",
			localityConfig: catpb.LocalityConfig{
				Locality: &catpb.LocalityConfig_RegionalByTable_{
					RegionalByTable: &catpb.LocalityConfig_RegionalByTable{
						Region: nil,
					},
				},
			},
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
				"region_d",
			}, "region_b", descpb.SurvivalGoal_ZONE_FAILURE, descpb.InvalidID, descpb.DataPlacement_DEFAULT, nil),
			expected: *(zonepb.NewZoneConfig()),
		},
		{
			desc: "4-region regional by table with regional survival on primary region",
			localityConfig: catpb.LocalityConfig{
				Locality: &catpb.LocalityConfig_RegionalByTable_{
					RegionalByTable: &catpb.LocalityConfig_RegionalByTable{
						Region: nil,
					},
				},
			},
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
				"region_d",
			}, "region_b", descpb.SurvivalGoal_REGION_FAILURE, descpb.InvalidID, descpb.DataPlacement_DEFAULT, nil),
			expected: *(zonepb.NewZoneConfig()),
		},
		{
			desc: "4-region regional by table with zone survival on non primary region",
			localityConfig: catpb.LocalityConfig{
				Locality: &catpb.LocalityConfig_RegionalByTable_{
					RegionalByTable: &catpb.LocalityConfig_RegionalByTable{
						Region: protoRegionName("region_c"),
					},
				},
			},
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
				"region_d",
			}, "region_b", descpb.SurvivalGoal_ZONE_FAILURE, descpb.InvalidID, descpb.DataPlacement_DEFAULT, nil),
			expected: zonepb.ZoneConfig{
				NumReplicas: nil, // Set at the database level.
				NumVoters:   proto.Int32(3),
				LeasePreferences: []zonepb.LeasePreference{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_c"},
						},
					},
				},
				InheritedConstraints:        true,
				NullVoterConstraintsIsEmpty: true,
				VoterConstraints: []zonepb.ConstraintsConjunction{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_c"},
						},
					},
				},
			},
		},
		{
			desc: "4-region regional by table with regional survival on non primary region",
			localityConfig: catpb.LocalityConfig{
				Locality: &catpb.LocalityConfig_RegionalByTable_{
					RegionalByTable: &catpb.LocalityConfig_RegionalByTable{
						Region: protoRegionName("region_c"),
					},
				},
			},
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
				"region_d",
			}, "region_b", descpb.SurvivalGoal_REGION_FAILURE, descpb.InvalidID, descpb.DataPlacement_DEFAULT, nil),
			expected: zonepb.ZoneConfig{
				NumReplicas: nil, // Set at the database level.
				NumVoters:   proto.Int32(5),
				LeasePreferences: []zonepb.LeasePreference{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_c"},
						},
					},
				},
				InheritedConstraints:        true,
				NullVoterConstraintsIsEmpty: true,
				VoterConstraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 2,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_c"},
						},
					},
				},
			},
		},
		{
			desc: "4-region global table with restricted placement",
			localityConfig: catpb.LocalityConfig{
				Locality: &catpb.LocalityConfig_Global_{},
			},
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
				"region_d",
			}, "region_b", descpb.SurvivalGoal_ZONE_FAILURE, descpb.InvalidID, descpb.DataPlacement_RESTRICTED, nil),
			expected: zonepb.ZoneConfig{
				NumReplicas:                 proto.Int32(6),
				NumVoters:                   proto.Int32(3),
				GlobalReads:                 proto.Bool(true),
				InheritedConstraints:        false,
				NullVoterConstraintsIsEmpty: true,
				InheritedLeasePreferences:   true,
				VoterConstraints: []zonepb.ConstraintsConjunction{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
				Constraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_c"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_d"},
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			zc, err := zoneConfigForMultiRegionTable(tc.localityConfig, tc.regionConfig)
			require.NoError(t, err)
			require.Equal(t, tc.expected, *zc)
		})
	}
}

func TestZoneConfigForMultiRegionPartition(t *testing.T) {
	defer leaktest.AfterTest(t)()

	testCases := []struct {
		desc         string
		region       catpb.RegionName
		regionConfig multiregion.RegionConfig
		expected     zonepb.ZoneConfig
	}{
		{
			desc:   "4-region table with zone survivability",
			region: "region_a",
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
				"region_d",
			}, "region_b", descpb.SurvivalGoal_ZONE_FAILURE, descpb.InvalidID, descpb.DataPlacement_DEFAULT, nil),
			expected: zonepb.ZoneConfig{
				NumReplicas:                 nil, // Set at the database level.
				NumVoters:                   proto.Int32(3),
				InheritedConstraints:        true,
				NullVoterConstraintsIsEmpty: true,
				VoterConstraints: []zonepb.ConstraintsConjunction{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
				},
				LeasePreferences: []zonepb.LeasePreference{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
				},
			},
		},
		{
			desc:   "4-region table with region survivability",
			region: "region_a",
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
				"region_d",
			}, "region_b", descpb.SurvivalGoal_REGION_FAILURE, descpb.InvalidID, descpb.DataPlacement_DEFAULT, nil),
			expected: zonepb.ZoneConfig{
				NumReplicas:                 nil, // Set at the database level.
				NumVoters:                   proto.Int32(5),
				InheritedConstraints:        true,
				NullVoterConstraintsIsEmpty: true,
				VoterConstraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 2,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
				},
				LeasePreferences: []zonepb.LeasePreference{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			zc, err := zoneConfigForMultiRegionPartition(tc.region, tc.regionConfig)
			require.NoError(t, err)
			require.Equal(t, tc.expected, zc)
		})
	}
}

func TestZoneConfigForRegionalByTableWithSuperRegions(t *testing.T) {
	defer leaktest.AfterTest(t)()

	const validMultiRegionEnumID = 100

	testCases := []struct {
		desc           string
		localityConfig catpb.LocalityConfig
		regionConfig   multiregion.RegionConfig
		expected       zonepb.ZoneConfig
	}{
		{
			desc: "super region with regional table, zone failure",
			localityConfig: catpb.LocalityConfig{
				Locality: &catpb.LocalityConfig_RegionalByTable_{
					RegionalByTable: &catpb.LocalityConfig_RegionalByTable{
						Region: protoRegionName("region_b"),
					},
				},
			},
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
				"region_d",
			}, "region_b", descpb.SurvivalGoal_ZONE_FAILURE, validMultiRegionEnumID, descpb.DataPlacement_DEFAULT, []descpb.SuperRegion{
				{
					SuperRegionName: "super_region_ab",
					Regions:         catpb.RegionNames{"region_a", "region_b"},
				},
			}),
			expected: zonepb.ZoneConfig{
				NumReplicas:                 proto.Int32(4),
				NumVoters:                   proto.Int32(3),
				InheritedConstraints:        false,
				NullVoterConstraintsIsEmpty: true,
				Constraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
				VoterConstraints: []zonepb.ConstraintsConjunction{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
				LeasePreferences: []zonepb.LeasePreference{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
			},
		},
		{
			desc: "super region with regional table region failure super region with 3 regions",
			localityConfig: catpb.LocalityConfig{
				Locality: &catpb.LocalityConfig_RegionalByTable_{
					RegionalByTable: &catpb.LocalityConfig_RegionalByTable{
						Region: protoRegionName("region_b"),
					},
				},
			},
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
				"region_d",
			}, "region_b", descpb.SurvivalGoal_REGION_FAILURE, validMultiRegionEnumID, descpb.DataPlacement_DEFAULT, []descpb.SuperRegion{
				{
					SuperRegionName: "super_region_abc",
					Regions:         catpb.RegionNames{"region_a", "region_b", "region_c"},
				},
			}),
			expected: zonepb.ZoneConfig{
				NumReplicas:                 proto.Int32(5),
				NumVoters:                   proto.Int32(5),
				InheritedConstraints:        false,
				NullVoterConstraintsIsEmpty: true,
				Constraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 2,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_c"},
						},
					},
				},
				VoterConstraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 2,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
				LeasePreferences: []zonepb.LeasePreference{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
			},
		},
		{
			desc: "super region with regional table region failure super region with 4 regions",
			localityConfig: catpb.LocalityConfig{
				Locality: &catpb.LocalityConfig_RegionalByTable_{
					RegionalByTable: &catpb.LocalityConfig_RegionalByTable{
						Region: protoRegionName("region_b"),
					},
				},
			},
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
				"region_d",
			}, "region_b", descpb.SurvivalGoal_REGION_FAILURE, validMultiRegionEnumID, descpb.DataPlacement_DEFAULT, []descpb.SuperRegion{
				{
					SuperRegionName: "super_region_abcd",
					Regions:         catpb.RegionNames{"region_a", "region_b", "region_c", "region_d"},
				},
			}),
			expected: zonepb.ZoneConfig{
				NumReplicas:                 proto.Int32(5),
				NumVoters:                   proto.Int32(5),
				InheritedConstraints:        false,
				NullVoterConstraintsIsEmpty: true,
				Constraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_c"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_d"},
						},
					},
				},
				VoterConstraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 2,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
				LeasePreferences: []zonepb.LeasePreference{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := multiregion.ValidateRegionConfig(tc.regionConfig)
			require.NoError(t, err)
			zc, err := zoneConfigForMultiRegionTable(tc.localityConfig, tc.regionConfig)
			require.NoError(t, err)
			require.Equal(t, tc.expected, *zc)
		})
	}
}

func TestZoneConfigForRegionalByRowPartitionsWithSuperRegions(t *testing.T) {
	defer leaktest.AfterTest(t)()

	const validMultiRegionEnumID = 100

	testCases := []struct {
		desc           string
		region         catpb.RegionName
		localityConfig catpb.LocalityConfig
		regionConfig   multiregion.RegionConfig
		expected       zonepb.ZoneConfig
	}{
		{
			desc: "super region with regional by row, zone failure, partition region_a",
			localityConfig: catpb.LocalityConfig{
				Locality: &catpb.LocalityConfig_RegionalByRow_{
					RegionalByRow: &catpb.LocalityConfig_RegionalByRow{},
				},
			},
			region: "region_a",
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
				"region_d",
			}, "region_b", descpb.SurvivalGoal_ZONE_FAILURE, validMultiRegionEnumID, descpb.DataPlacement_DEFAULT, []descpb.SuperRegion{
				{
					SuperRegionName: "super_region_ab",
					Regions:         catpb.RegionNames{"region_a", "region_b"},
				},
			}),
			expected: zonepb.ZoneConfig{
				NumReplicas:                 proto.Int32(4),
				NumVoters:                   proto.Int32(3),
				InheritedConstraints:        false,
				NullVoterConstraintsIsEmpty: true,
				Constraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
				VoterConstraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 0,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
				},
				LeasePreferences: []zonepb.LeasePreference{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
				},
			},
		},
		{
			desc: "super region with regional by row, zone failure, partition region_b",
			localityConfig: catpb.LocalityConfig{
				Locality: &catpb.LocalityConfig_RegionalByRow_{
					RegionalByRow: &catpb.LocalityConfig_RegionalByRow{},
				},
			},
			region: "region_b",
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
				"region_d",
			}, "region_b", descpb.SurvivalGoal_ZONE_FAILURE, validMultiRegionEnumID, descpb.DataPlacement_DEFAULT, []descpb.SuperRegion{
				{
					SuperRegionName: "super_region_ab",
					Regions:         catpb.RegionNames{"region_a", "region_b"},
				},
			}),
			expected: zonepb.ZoneConfig{
				NumReplicas:                 proto.Int32(4),
				NumVoters:                   proto.Int32(3),
				InheritedConstraints:        false,
				NullVoterConstraintsIsEmpty: true,
				Constraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
				VoterConstraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 0,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
				LeasePreferences: []zonepb.LeasePreference{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
			},
		},
		{
			desc: "super region with regional by row, region failure, partition region_a",
			localityConfig: catpb.LocalityConfig{
				Locality: &catpb.LocalityConfig_RegionalByRow_{
					RegionalByRow: &catpb.LocalityConfig_RegionalByRow{},
				},
			},
			region: "region_a",
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
				"region_d",
			}, "region_b", descpb.SurvivalGoal_REGION_FAILURE, validMultiRegionEnumID, descpb.DataPlacement_DEFAULT, []descpb.SuperRegion{
				{
					SuperRegionName: "super_region_abc",
					Regions:         catpb.RegionNames{"region_a", "region_b", "region_c"},
				},
			}),
			expected: zonepb.ZoneConfig{
				NumReplicas:                 proto.Int32(5),
				NumVoters:                   proto.Int32(5),
				InheritedConstraints:        false,
				NullVoterConstraintsIsEmpty: true,
				Constraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
					{
						NumReplicas: 2,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_c"},
						},
					},
				},
				VoterConstraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 2,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
				},
				LeasePreferences: []zonepb.LeasePreference{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
				},
			},
		},
		{
			desc: "super region with regional by row, region failure, partition region_b",
			localityConfig: catpb.LocalityConfig{
				Locality: &catpb.LocalityConfig_RegionalByRow_{
					RegionalByRow: &catpb.LocalityConfig_RegionalByRow{},
				},
			},
			region: "region_b",
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
				"region_d",
			}, "region_b", descpb.SurvivalGoal_REGION_FAILURE, validMultiRegionEnumID, descpb.DataPlacement_DEFAULT, []descpb.SuperRegion{
				{
					SuperRegionName: "super_region_abc",
					Regions:         catpb.RegionNames{"region_a", "region_b", "region_c"},
				},
			}),
			expected: zonepb.ZoneConfig{
				NumReplicas:                 proto.Int32(5),
				NumVoters:                   proto.Int32(5),
				InheritedConstraints:        false,
				NullVoterConstraintsIsEmpty: true,
				Constraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 2,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_c"},
						},
					},
				},
				VoterConstraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 2,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
				LeasePreferences: []zonepb.LeasePreference{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
			},
		},
		{
			desc: "super region with regional by row, region failure, 6 regions, super region with 5 regions",
			localityConfig: catpb.LocalityConfig{
				Locality: &catpb.LocalityConfig_RegionalByRow_{
					RegionalByRow: &catpb.LocalityConfig_RegionalByRow{},
				},
			},
			region: "region_b",
			regionConfig: multiregion.MakeRegionConfig(catpb.RegionNames{
				"region_b",
				"region_c",
				"region_a",
				"region_d",
				"region_e",
				"region_f",
			}, "region_b", descpb.SurvivalGoal_REGION_FAILURE, validMultiRegionEnumID, descpb.DataPlacement_DEFAULT, []descpb.SuperRegion{
				{
					SuperRegionName: "super_region_abcde",
					Regions:         catpb.RegionNames{"region_a", "region_b", "region_c", "region_d", "region_e"},
				},
			}),
			expected: zonepb.ZoneConfig{
				NumReplicas:                 proto.Int32(6),
				NumVoters:                   proto.Int32(5),
				InheritedConstraints:        false,
				NullVoterConstraintsIsEmpty: true,
				Constraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_a"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_c"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_d"},
						},
					},
					{
						NumReplicas: 1,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_e"},
						},
					},
				},
				VoterConstraints: []zonepb.ConstraintsConjunction{
					{
						NumReplicas: 2,
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
				LeasePreferences: []zonepb.LeasePreference{
					{
						Constraints: []zonepb.Constraint{
							{Type: zonepb.Constraint_REQUIRED, Key: "region", Value: "region_b"},
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := multiregion.ValidateRegionConfig(tc.regionConfig)
			require.NoError(t, err)
			zc, err := zoneConfigForMultiRegionPartition(tc.region, tc.regionConfig)
			require.NoError(t, err)
			require.Equal(t, tc.expected, zc)
		})
	}
}

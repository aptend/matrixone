// Copyright 2026 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plan

import (
	"context"
	"testing"

	"github.com/matrixorigin/matrixone/pkg/container/types"
	planpb "github.com/matrixorigin/matrixone/pkg/pb/plan"
	"github.com/stretchr/testify/require"
)

func TestSetMembershipJoinPublicPlan(t *testing.T) {
	const join = " from nation n join region r on n.n_regionkey = r.r_regionkey"
	for _, tc := range []struct {
		name string
		sql  string
		semi int
	}{
		{"union_left", "select n.n_nationkey" + join + " union select n_nationkey from nation", 1},
		{"intersect_right", "select r.r_regionkey" + join + " intersect select r_regionkey from region", 1},
		{"except_right", "select r.r_regionkey" + join + " except select r_regionkey from region", 1},
		{"all_observes_count", "select n.n_nationkey" + join + " union all select n_nationkey from nation", 0},
		{"mixed_outputs", "select n.n_nationkey, r.r_regionkey" + join + " union select n_nationkey, n_regionkey from nation", 0},
		{"limited_arm", "(select n.n_nationkey" + join + " limit 1) union select n_nationkey from nation", 0},
		{"aggregate_observes_count", "select count(*)" + join + " union select count(*) from nation", 0},
		{"window_observes_count", "select row_number() over(order by n.n_nationkey)" + join + " union select n_nationkey from nation", 0},
		{"fallible_output", "select cast(n.n_name as bigint)" + join + " union select n_nationkey from nation", 0},
		{"fallible_origin", "select n.n_nationkey from nation n join (select cast(r_name as bigint) k from region) r on n.n_regionkey = r.k union select n_nationkey from nation", 0},
		{"volatile_output", "select rand()" + join + " union select rand() from nation", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := NewMockCompilerContext(true, newPlanTestProcess(t))
			p, err := buildOneQuery(t, ctx, tc.sql, false)
			require.NoError(t, err)
			query := p.GetQuery()
			require.NotNil(t, query)
			semi := 0
			seen := make(map[int32]bool)
			var visit func(int32)
			visit = func(id int32) {
				if seen[id] {
					return
				}
				seen[id] = true
				node := query.Nodes[id]
				if node.NodeType == planpb.Node_JOIN && node.JoinType == planpb.Node_SEMI {
					semi++
				}
				for _, child := range node.Children {
					visit(child)
				}
			}
			for _, root := range query.Steps {
				visit(root)
			}
			require.Equal(t, tc.semi, semi)
		})
	}
}

func TestSetMembershipJoinOwnershipAndDependencies(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*QueryBuilder)
		semi   bool
	}{
		{"current_root_not_stale_step", func(b *QueryBuilder) { b.qry.Steps[0] = 3 }, true},
		{"shared_project", func(b *QueryBuilder) {
			b.qry.Nodes = append(b.qry.Nodes, &planpb.Node{NodeType: planpb.Node_AGG, Children: []int32{3}})
			b.qry.Steps = append(b.qry.Steps, 7)
		}, false},
		{"runtime_filter_dependency", func(b *QueryBuilder) {
			b.qry.Nodes[2].RuntimeFilterBuildList = []*planpb.RuntimeFilterSpec{{}}
		}, false},
		{"materialized_input", func(b *QueryBuilder) { b.qry.Nodes[1].NodeType = planpb.Node_SINK_SCAN }, false},
		{"predicate_observes_discarded_side", func(b *QueryBuilder) {
			b.qry.Nodes[2].FilterList = b.qry.Nodes[2].OnList
		}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := newSetMembershipTestBuilder(t)
			tc.change(b)
			projection := DeepCopyExpr(b.qry.Nodes[3].ProjectList[0])
			b.reduceSetMembershipJoins(6, 0)
			want := planpb.Node_INNER
			if tc.semi {
				want = planpb.Node_SEMI
				require.Equal(t, []int32{1, 0}, b.qry.Nodes[2].Children)
				require.False(t, b.qry.Nodes[2].IsRightJoin)
			}
			require.Equal(t, want, b.qry.Nodes[2].JoinType)
			require.Equal(t, projection, b.qry.Nodes[3].ProjectList[0])
			// The rule does not revisit or reorient an already reduced join.
			b.reduceSetMembershipJoins(6, 0)
			require.Equal(t, want, b.qry.Nodes[2].JoinType)
		})
	}
}

func newSetMembershipTestBuilder(t *testing.T) *QueryBuilder {
	t.Helper()
	typ := planpb.Type{Id: int32(types.T_int64)}
	scan := func(tag int32) *planpb.Node {
		return &planpb.Node{NodeType: planpb.Node_TABLE_SCAN, BindingTags: []int32{tag},
			TableDef: &planpb.TableDef{Cols: []*planpb.ColDef{{Name: "k", Typ: typ}}}}
	}
	eq, err := BindFuncExprImplByPlanExpr(context.Background(), "=", []*planpb.Expr{
		GetColExpr(typ, 10, 0), GetColExpr(typ, 11, 0),
	})
	require.NoError(t, err)
	return &QueryBuilder{qry: &planpb.Query{Steps: []int32{6}, Nodes: []*planpb.Node{
		scan(10), scan(11),
		{NodeType: planpb.Node_JOIN, JoinType: planpb.Node_INNER, Children: []int32{0, 1}, OnList: []*planpb.Expr{eq}},
		{NodeType: planpb.Node_PROJECT, Children: []int32{2}, BindingTags: []int32{12}, ProjectList: []*planpb.Expr{GetColExpr(typ, 11, 0)}},
		scan(13),
		{NodeType: planpb.Node_PROJECT, Children: []int32{4}, BindingTags: []int32{14}, ProjectList: []*planpb.Expr{GetColExpr(typ, 13, 0)}},
		{NodeType: planpb.Node_UNION, Children: []int32{3, 5}, BindingTags: []int32{15}},
	}}}
}

package node

import (
	"context"
	"fmt"

	"github.com/expr-lang/expr"
	wf "github.com/showntop/llmack/workflow"
)

type (
	// exclGateway is an exclusive gateway that can return exactly one branch
	exclGateway struct {
		executeable
		Identifier
		outging []*wf.Edge
	}

	// forkGateway is an fork gateway that can return all branch
	forkGateway struct {
		executeable
		Identifier
		outging []*wf.Edge
	}
)

// ExclGateway fn initializes exclusive gateway
// func ExclGateway(pp ...*GatewayCondition) (*exclGateway, error) {
func ExclGateway(node *wf.Node, outgoing ...*wf.Edge) (*exclGateway, error) {
	t := len(outgoing)
	if t < 2 {
		return nil, fmt.Errorf("expecting at least two branches for exclusive gateway")
	}

	return &exclGateway{outging: outgoing}, nil
}

// Kind TODO
func (gw exclGateway) Kind() wf.NodeKind {
	return wf.NodeKindGateway
}

// GatewayActivity TODO
func GatewayActivity(ctx context.Context, n *exclGateway, r *ExecRequest) (ExecResponse, error) {
	return n.Execute(ctx, r)
}

// Exec fn on exclGateway uses current scope to test all configured conditions
//
// Exactly one matched path can be returned.
func (gw exclGateway) Execute(ctx context.Context, r *ExecRequest) (ExecResponse, error) {
	test := func(ctx context.Context, vars map[string]any, express string) (bool, error) {
		out, err := expr.Eval(express, vars)
		if err != nil {
			return false, err
		}

		return out.(bool), nil
	}
	_ = test
	for _, p := range gw.outging {
		expr := p.Express
		if expr == "" {
			// empty & last; treat it as else part of the if condition
			return wf.NodeID{ID: p.Target}, nil
		}

		if result, err := test(ctx, r.Inputs, expr); err != nil {
			return nil, err
		} else if result {
			return wf.NodeID{ID: p.Target}, nil
		}
	}

	return nil, fmt.Errorf("exclusive gateway must match one condition")
}

// ForkGateway fn initializes fork gateway
func ForkGateway(node *wf.Node, outgoing ...*wf.Edge) (*forkGateway, error) {
	t := len(outgoing)
	if t < 2 {
		return nil, fmt.Errorf("expecting at least two branches for fork gateway")
	}

	return &forkGateway{outging: outgoing}, nil
}

// Kind TODO
func (gw forkGateway) Kind() wf.NodeKind {
	return wf.NodeKindGateway
}

// ForkGatewayActivity TODO
func ForkGatewayActivity(ctx context.Context, n *forkGateway, r *ExecRequest) (ExecResponse, error) {
	return n.Execute(ctx, r)
}

// Exec fn on forkGateway uses current scope to test all configured conditions
//
// All path can be returned.
func (gw forkGateway) Execute(ctx context.Context, r *ExecRequest) (ExecResponse, error) {
	// create approval apply

	// hack ? 需求太具体
	// newctx := wfctx.StandandActivity(ctx)

	// apply := makeApprovalApply(r)
	// apply.Type = gw.
	// if v, err := r.Scope.Select(Var_Name_Process_RunnerTerritory); err == nil && v != nil {
	// 	apply.BelongTerritory = fmt.Sprint(v.Get())
	// } else {
	// 	territory, err := dao.QueryUserTerritory(newctx, 0, apply.Starter)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	apply.BelongTerritory = territory.Code
	// }

	// if err := apply.CreateApprovalApply(newctx); err != nil {
	// 	return nil, err
	// }

	return nil, nil
	// for _, p := range gw.outging {
	// expr := p.Expr()
	// if expr == "" {
	// 	// empty & last; treat it as else part of the if condition
	// 	// return r.Graph.NodeByID(p.Target), nil
	// 	return wf.NodeID(p.Target), nil
	// }

	// if result, err := test(ctx, r.Scope, expr); err != nil {
	// 	return nil, err
	// } else if result {
	// 	return wf.NodeID(p.Target), nil
	// }
	// }

	// return nil, fmt.Errorf("exclusive gateway must match one condition")
}

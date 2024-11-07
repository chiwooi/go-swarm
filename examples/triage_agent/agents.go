package main

import (
    "fmt"

    "github.com/chiwooi/go-swarm"
    "github.com/chiwooi/go-swarm/option"
    "github.com/chiwooi/go-swarm/types"
)

type ProcessRefundArgs struct {
    ItemID string `json:"item_id" desc:"The item ID to refund." required:"true"`
    Reason string `json:"reason"  desc:"The reason for the refund."`
}

func ProcessRefund(ctx goswarm.Context, args ProcessRefundArgs) string {
    if ctx.IsAnalyze() {
        ctx.SetDescription("Refund an item. Make sure you have the item_id of the form item_... Ask for user confirmation before processing the refund.")
        return ""
    }

    if args.Reason == "" {
        args.Reason = "NOT SPECIFIED"
    }

    fmt.Printf("[mock] Refunding item %s because %s...\n", args.ItemID, args.Reason)
    return "Success!"
}

func ApplyDiscount(ctx goswarm.Context) string {
    if ctx.IsAnalyze() {
        ctx.SetDescription("Apply a discount to the user's cart.")
        return ""
    }

    fmt.Println("[mock] Applying discount...")
    return "Applied discount of 11%"
}

var triageAgent = goswarm.NewAgent(
    option.WithAgentName("Triage Agent"), 
    option.WithAgentInstructions("Determine which agent is best suited to handle the user's request, and transfer the conversation to that agent."),
)
var salesAgent = goswarm.NewAgent(
    option.WithAgentName("Sales Agent"), 
    option.WithAgentInstructions("Be super enthusiastic about selling bees."),
)
var refundsAgent = goswarm.NewAgent(
    option.WithAgentName("Refunds Agent"), 
    option.WithAgentInstructions("Help the user with a refund. If the reason is that it was too expensive, offer the user a refund code. If they insist, then process the refund."),
    option.WithAgentFunctions(ProcessRefund, ApplyDiscount),
)

func transferBackToTriage(ctx goswarm.Context) *types.Agent {
    if ctx.IsAnalyze() {
        ctx.SetDescription("Call this function if a user is asking about a topic that is not handled by the current agent.")
        return nil
    }
    return triageAgent
}

func transferToSales(ctx goswarm.Context) *types.Agent {
    return salesAgent
}

func transferToRefunds(ctx goswarm.Context) *types.Agent {
    return refundsAgent
}

func init() {
    triageAgent.Functions = append(triageAgent.Functions, transferToSales, transferToRefunds)
    salesAgent.Functions = append(salesAgent.Functions, transferBackToTriage)
    refundsAgent.Functions = append(refundsAgent.Functions, transferBackToTriage)
}

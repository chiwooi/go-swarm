// https://github.com/openai/swarm/blob/main/examples/basic/function_calling.py
package main

import (
    "fmt"
    "time"

    "github.com/openai/openai-go"

    "github.com/chiwooi/go-swarm"
    "github.com/chiwooi/go-swarm/option"
)

type GetWeatherArgs struct {
    Location string    `json:"location" desc:"The location to get the weather for." required:"true"`
    Time     time.Time `json:"time"     desc:"The time to get the weather for."`
}

func GetWeather(ctx goswarm.Context, args GetWeatherArgs) string {
    if ctx.IsAnalyze() {
        ctx.SetDescription("Get the current weather in a given location. Location MUST be a city.")
        return ""
    }

    if args.Time.IsZero() {
        args.Time = time.Now()
    }

    return fmt.Sprintf("{\"location\": \"%s\", \"temperature\":\"65\", \"time\": \"%s\"}", args.Location, args.Time.Format(time.RFC3339))
}

type SendEmailArgs struct {
    Recipient string `json:"recipient" desc:"The email address to send the email to." required:"true"`
    Subject   string `json:"subject"   desc:"The subject of the email." required:"true"`
    Body      string `json:"body"      desc:"The body of the email." required:"true"`
}

func SendEmail(ctx goswarm.Context, args SendEmailArgs) string {
    if ctx.IsAnalyze() {
        ctx.SetDescription("Send an email to a recipient.")
        return ""
    }

    fmt.Println("Sending email...")
    fmt.Printf("To: %s\n", args.Recipient)
    fmt.Printf("Subject: %s\n", args.Subject)
    fmt.Printf("Body: %s\n", args.Body)

    return "Sent!"
}

var oai = openai.NewClient()
var client = goswarm.NewSwarm(oai)

var weatherAgent = goswarm.NewAgent("Weather Agent",
    option.WithAgentInstructions("You are a helpful agent."),
    option.WithAgentToolChoice(option.ToolChoiceOptionAuto),
    option.WithAgentFunctions(GetWeather, SendEmail),
)


package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"

	"ai-staffing/internal/notify/gmail"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	usage := errors.New("dùng authorize -client FILE -from EMAIL, check, hoặc send-check -to EMAIL")
	if len(args) == 0 {
		return usage
	}
	switch args[0] {
	case "authorize":
		flags := flag.NewFlagSet("authorize", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		client := flags.String("client", "", "file OAuth Desktop client")
		from := flags.String("from", "", "hộp thư gửi")
		if flags.Parse(args[1:]) != nil || flags.NArg() != 0 {
			return usage
		}
		if err := gmail.Authorize(ctx, *client, *from, func(url string) {
			fmt.Println("Mở URL sau trong trình duyệt của máy này và chọn đúng hộp thư gửi:")
			fmt.Println(url)
		}); err != nil {
			return err
		}
		fmt.Println("Đã lưu cấu hình OAuth Gmail.")
	case "check":
		if len(args) != 1 {
			return usage
		}
		sender, err := gmail.NewFromEnv()
		if err != nil {
			return err
		}
		if err := sender.CheckConnection(ctx); err != nil {
			return err
		}
		fmt.Println("Refresh access token Gmail thành công.")
	case "send-check":
		flags := flag.NewFlagSet("send-check", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		to := flags.String("to", "", "hộp thư nhận đã được phép")
		if flags.Parse(args[1:]) != nil || flags.NArg() != 0 || *to == "" {
			return usage
		}
		sender, err := gmail.NewFromEnv()
		if err != nil {
			return err
		}
		status, err := sender.SendCheck(ctx, *to)
		fmt.Println(status)
		return err
	default:
		return usage
	}
	return nil
}

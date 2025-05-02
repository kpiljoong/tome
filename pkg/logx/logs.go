package logx

import "fmt"

func Success(msg string, args ...any) {
	fmt.Printf("✅ "+msg+"\n", args...)
}

func Info(msg string, args ...any) {
	fmt.Printf("📘 "+msg+"\n", args...)
}

func Warn(msg string, args ...any) {
	fmt.Printf("⚠️  "+msg+"\n", args...)
}

func Error(msg string, args ...any) {
	fmt.Printf("❌ "+msg+"\n", args...)
}

func Hint(msg string, args ...any) {
	fmt.Printf("💡 "+msg+"\n", args...)
}

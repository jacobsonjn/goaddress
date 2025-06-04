# Project Setup
```go
mkdir golangcliscaffold && cd golangcliscaffold
go mod init github.com/mwiater/golangcliscaffold
go install github.com/spf13/cobra-cli@latest
cobra-cli init
```

This will create the following project setup:
```bash
.
├── cmd
│   └── root.go
├── go.mod
├── go.sum
├── LICENSE
└──  main.go

1 directory, 5 files
```
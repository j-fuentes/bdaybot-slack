//go:generate protoc -I. --go_out=paths=source_relative:. config.proto bday.proto

package bdaybot

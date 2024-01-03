package iproto

//go:generate ./generate.sh
//go:generate stringer -type=Error
//go:generate stringer -type=Feature
//go:generate stringer -type=Flag
//go:generate stringer -type=Iterator
//go:generate stringer -type=Type
//go:generate stringer -type=Key,MetadataKey,BallotKey,RaftKey,SqlInfoKey -output=keys_string.go
//go:generate goimports -w .

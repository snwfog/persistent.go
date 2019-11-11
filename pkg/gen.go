package pkg

//go:generate genny -in=./generic/linked_list.go -out=../pkg/int/linked_list.go -pkg=int gen "Value=int"
//go:generate genny -in=./generic/linked_list.go -out=../pkg/campaign/linked_list.go -pkg=campaign gen "Value=Campaign"

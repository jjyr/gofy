package main

func foo() {
	for {}
}

func main() {
	go foo()
	for {}
}

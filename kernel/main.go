package main

func main() {
	initconsole()
	initframes()
	initpaging()
	initinterrupts()
	puts("GOFY\n")
	initproc()
	a := procreate()
	putnum(a, 10)
	swtch()
}

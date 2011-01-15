package runtime

func fuck(s string)

func panicslice() {
	fuck("Slice fucked up")
}

func throwreturn() {
	fuck("Some seriously bad compiler shit happened (throwreturn)")
}

func panicindex() {
	fuck("Bounds check failed")
}

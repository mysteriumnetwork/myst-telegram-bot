package main

func main() {
	keystore.GetKeystore().ImportECDSA("fbdbd0415ee7a937411a42ee356a203023d792d802b0a83b181b2d2b9f58a4f6", "Opa2u3223")
	keystore.GetKeystore().Import()
}

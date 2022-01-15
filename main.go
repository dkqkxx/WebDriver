package main

func main() {
	clawler := NewCrawler("./driver/chromedriver.exe", 8080, 1)
	clawler.Scratch("https://fullstack.love", nil)
}

build_pi:
	env GOOS=linux GOARCH=arm GOARM=5 go build -o flarm_arm5
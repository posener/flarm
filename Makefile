build_pi:
	env GOOS=linux GOARCH=arm GOARM=5 go build -o flarm_arm5

build:
	go build -o flarm

pack: build_pi build
	rm flarm.zip
	zip -r flarm.zip cesium flarm flarm_arm5 config.json
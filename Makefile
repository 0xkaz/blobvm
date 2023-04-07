_NAME := blobvm

buildsh: kill
	# rm -rf ./build/*
	./scripts/build.sh
	# ./scripts/run2-pre.sh1.7.13
run: kill
	# ./scripts/run.sh 1.7.13
	./scripts/run.sh 1.9.3
	# ./scripts/run.sh 1.9.10
	
	# ./scripts/run2.sh 1.7.13

run2: kill

	./scripts/run-without-anr.sh 1.9.3

kill:
	# ps awxww | grep avalanche | awk '{print $$1}' | xargs  kill -9 
	ps awxww | grep avalanche | grep -v grep | awk '{print $$1}'|xargs  kill -9 
ps:
	ps awxww | grep avalanche 

docker: 
	docker build   -t $(_NAME) . 

dockerrun: 
	docker run -p 12352:12352 -t $(_NAME)  

console: 
	docker run -p 12352:12352 -it $(_NAME)  /bin/bash 


# fmt:
# 	./scripts/gofmt-w.sh



_NAME := blobvm

buildsh: kill
	# rm -rf ./build/*
	./scripts/build.sh
	./scripts/run2-pre.sh 1.9.7
run: kill
	# ./scripts/run.sh 1.9.7
	./scripts/run.sh 1.9.7
	# ./scripts/run2.sh 1.9.7
kill:
	# ps awxww | grep avalanche | awk '{print $$1}' | xargs  kill -9 
	ps awxww | grep avalanche | grep -v grep | awk '{print $$1}'|xargs  kill -9 
ps:
	ps awxww | grep avalanche 

dockerbuild: 
	docker build   -t $(_NAME) . 

dockerrun: 
	docker run -p 12352:12352 -t $(_NAME)  

console: 
	docker run -p 12352:12352 -it $(_NAME)  /bin/bash 


# fmt:
# 	./scripts/gofmt-w.sh



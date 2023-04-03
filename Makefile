_NAME := blobvm

buildsh:
	rm -rf ./build/*
	./scripts/build.sh
run: 
	# ./scripts/run.sh 1.9.7
	./scripts/run2.sh 1.9.7
kill:
	ps awxww | grep avalanche | awk '{print $$1}' | xargs  kill -9 &>2
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



 **To build docker image:**  
 ```docker build -t volley-polley .```  
 **To start:**  
 Put private.key and .env files to directory and run  
 ```docker run --rm -it -v "$(pwd)":/go/src/app volley-polley```
## How to run  
```go run main.go```  

## Share-X custom upload settings  
Request type: ```POST```  
Request url: ```http://localhost:8080/upload```  
File form name: ```uploadfile```  

Response type: ```Response text```  
Regex from response: ```"path":"(.+)"```  
URL: ```http://localhost:8080/$1,1$```  

### agentd

Nitrous.IO agent that runs on boxhost and processes requests to containers

### Endpoints

```
- GET "/containers/{id:.*}/file": read file from container
- GET "/containers/{id:.*}/dir": list directory entries from container
- PUT "/containers/{id:.*}/file": write content to a file
- PUT "/containers/{id:.*}/rename": rename a path
- POST "/containers/{id:.*}/file": write a new file
- POST "/containers/{id:.*}/dir": create a new directory
- POST "/containers/{id:.*}/attach": attach a shell to the container (http stream) and return the session id
- POST "/shell/{id:.*}/resize": resize a running shell
- POST	"/containers/{id:.*}/command": run a command in  the container
- DELETE "/containers/{id:.*}/path": delete a path in the container
```

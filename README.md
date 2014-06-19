### containerd

Nitrous.IO agent that runs on boxhost and processes requests to containers. It
does:
- expose commands to modify the container filesystem
- run commands inside the container
- attach a shell to the container

### Endpoints

```
- GET "/containers/{id:.*}/file": read file from container
- GET "/containers/{id:.*}/dir": list directory entries from container
- PUT "/containers/{id:.*}/file": write content to a file
- PUT "/containers/{id:.*}/rename": rename a path
- POST "/containers/{id:.*}/file": write a new file
- POST "/containers/{id:.*}/dir": create a new directory
- POST "/containers/{id:.*}/shell": register the intention to create a shell, this will return a shell id that we tracks
- POST "/shell/{id:.*}/attach": attach to the shell with this id to the container (http stream), this will actually start the shell
- POST "/shell/{id:.*}/resize": resize a running shell
- POST "/containers/{id:.*}/command": run a command in  the container
- DELETE "/containers/{id:.*}/path": delete a path in the container
```

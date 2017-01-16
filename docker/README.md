Banshee Docker Image
====
## Run banshee image

Start banshee image binding the external port `2015` for detector and port `2016` for webapp
```
    docker run -p 2015:2015 -p 2016:2016 eleme/banshee
```
## Use your own configuration file and persist storage
1. Create Data directory 

   ```
   mkdir data
   ```
2. Put `config.yaml` into `data` directory
2. Run banshee docker image

   ```
   docker run -p 2015:2015 -p 2016:2016 -v `pwd`/data:/banshee/data eleme/banshee
   ``` 

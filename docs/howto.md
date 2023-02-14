# How to work with template
go-service-template is fully functional service. You can build, run and so on. 

Далее описывается как из шаблона сделать новый сервис.  

1. Create empty repository for new service (fo example - myservice)
2. Clone go-service-template
3. Rename folder (go-service-template -> myservice) 
> mv go-service-template myservice

4. Replace remote repository
>   git remote rm origin
>   git remote add origin git@myservice

5. Rename module name. Script will change go.mod, .gitlab-ci.yaml and import paths in  all go files.   
>  ./scripts/renameModule.sh  "myservice"

6. Check dependencies
>   go mod tidy
>   go mod verify

7. Prepare database 
- create database (./scripts/01.database.sql)
- create users and schema (./scripts/02.schema.sql)
8. Generate swagger docs. Generated files (../../docs/swagger) must be added to repository. 
> cd ./internal/app
> swag init  --output ../../docs/swagger
9. Check and fix TODO
10. git push 

OLD_MODULE="go-service-template"
NEW_MODULE=$1

go mod edit -module $NEW_MODULE
#-- rename all imported module

find ./../ -type f -name '*.go' \
  -exec sed -i -e "s,$OLD_MODULE,$NEW_MODULE,g" {} \;

find ./../ -type f -name '*.gitlab-ci.yml' \
  -exec sed -i -e "s,$OLD_MODULE,$NEW_MODULE,g" {} \;
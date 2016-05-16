#!/usr/bin/env bash 

set -e 
 
repo_path="github.com/warebot/prometheusproxy" 
 
version=$( cat version/VERSION ) 
revision=$( git rev-parse --short HEAD 2> /dev/null || echo 'unknown' ) 
branch=$( git rev-parse --abbrev-ref HEAD 2> /dev/null || echo 'unknown' ) 
host=$( hostname -f ) 
build_date=$( date +%Y%m%d-%H:%M:%S ) 
go_version=$( go version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/' ) 

if [ "$(go env goos)" = "windows" ]; then 
        ext=".exe" 
fi 
 
ldflags=" 
  -X ${repo_path}/version.Version=${version} 
  -X ${repo_path}/version.Revision=${revision} 
  -X ${repo_path}/version.Branch=${branch} 
  -X ${repo_path}/version.BuildUser=${user}@${host} 
  -X ${repo_path}/version.BuildDate=${build_date} 
  -X ${repo_path}/version.GoVersion=${go_version}" 
 
export go15vendorexperiment="1" 
echo $ldflags 
#exit(1) 

echo " >   fetching dependencies"
go get


echo " >   running tests"
go test -v

echo " >   building"

echo " >   prometheusproxy" 
go build -ldflags "${ldflags}" -o bin/prometheusproxy_$version github.com/warebot/prometheusproxy/app

exit 0 



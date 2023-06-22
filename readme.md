# rexpander - Expands the ARNs

__Input__ : res_strings.txt
A file with ARNs to be expanded, each on a new line. Change the arns in res_strings.txt with the ARNs to be expanded.

__Output__ : set of expanded ARNs 

### Functionalities
__Services supported__ : S3, Lambda, Dynamobd, cloudtrail, redshift
__Wildcards supported__ : *, ?



### How to execute?
download sitools at same level as entry and resource_expander
make necessary changes in const to configure your local database connection at bottom inside the resource_expander/expander.go
1. open terminal
2. move to entry/
3. go run .

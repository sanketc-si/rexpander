package rexpander

import (
	"fmt"
	"strings"

	"database/sql"

	_ "github.com/lib/pq"
	"github.com/stackidentity/sitools/pkg/awsutil"
)

var db *sql.DB

type Set map[string]struct{}

var allPartitions = [...]string{"aws", "aws-us-gov", "aws-cn", "aws-cn-northwest-1"}

// var allRegions = [...]string{}

func handleStarWildcard(currentGeneratedResources []string, arnField string) []string {
	resourcesArray := []string{}

	switch arnField {
	case "Partition":
		for _, partition := range allPartitions {
			for _, currentResourceString := range currentGeneratedResources {
				currentResourceString = currentResourceString + ":" + partition
				resourcesArray = append(resourcesArray, currentResourceString)
			}

		}
	case "Service":
		// fetch every sservice from db
		allServices := []string{"lambda", "s3", "dynamodb", "cloudtrail"}
		for _, service := range allServices {
			for _, currentResourceString := range currentGeneratedResources {
				currentResourceString = currentResourceString + ":" + service
				resourcesArray = append(resourcesArray, currentResourceString)
			}

		}
	case "Region":
		//  get regions from global infrastructure or db
		allRegions := []string{"us-east-2", "us-east-1", "us-west-1", "us-west-2", "ap-east-1"}
		for _, region := range allRegions {
			for _, currentResourceString := range currentGeneratedResources {
				currentResourceString = currentResourceString + ":" + region
				resourcesArray = append(resourcesArray, currentResourceString)
			}

		}
	case "AccountId":
		//  get regions from global infrastructure or db
		allAccountId := []string{"#1", "#2", "#3", "#4"}
		for _, accountid := range allAccountId {
			for _, currentResourceString := range currentGeneratedResources {
				currentResourceString = currentResourceString + ":" + accountid
				resourcesArray = append(resourcesArray, currentResourceString)
			}

		}

	default:
		// code to execute when none of the values match
		fmt.Println("Unknown field")
	}

	return resourcesArray
}
func replaceWildcards(clause string) string {
	clause = strings.Replace(clause, "?", "_", -1)
	clause = strings.Replace(clause, "*", "%", -1)
	return clause
}

func getResourceTypes(resourceTypes map[string][]string, query string, res string) map[string][]string {

	rows, err := db.Query(query)
	if err != nil {
		// handle error

		panic(err)

	}
	defer rows.Close()
	for rows.Next() {
		var column1Type string
		err := rows.Scan(&column1Type)
		if err != nil {
			// handle error
			panic(err)
			// fmt.Println("at 141")
		}
		// Process the retrieved data
		// fmt.Println(column1Type)
		resourceTypes[res] = append(resourceTypes[res], column1Type)
	}
	if err := rows.Err(); err != nil {
		// handle error
		// fmt.Println("at 150")
		panic(err)

	}
	defer rows.Close()
	return resourceTypes
}
func fetchFromDb(resourcePathArray []string, query string) []string {

	if len(resourcePathArray) == 0 {

		rows, err := db.Query(query)
		if err != nil {

			panic(err)

		}
		defer rows.Close()
		for rows.Next() {
			var column1Type string
			err := rows.Scan(&column1Type)
			if err != nil {

				panic(err)

			}

			resourcePathArray = append(resourcePathArray, column1Type)
		}
		if err := rows.Err(); err != nil {

			panic(err)

		}
		defer rows.Close()
	}
	return resourcePathArray
}

// func handle handleClause(resource string, ) []string{

// }

func handleResourcesField(arnType *awsutil.Arn, currentGeneratedResources []string) []string {
	// if the resource is specified in the arn
	var resourceFieldExpanded = []string{}

	for i := range currentGeneratedResources {
		// arn:aws:s3:us-east-1:123
		arn := strings.SplitN(currentGeneratedResources[i], ":", 5)
		// arn, err := awsutil.ParseArn(currentGeneratedResources[i])
		if arn != nil {
			// service
			switch arn[2] {
			// arn:aws:s3:us-east-1:123:bucket
			case "s3":
				// resource is given
				if arnType.Resource != "*" && arnType.Resource != "" && !strings.ContainsAny(arnType.Resource, "*?[]{}^") {
					for i := range currentGeneratedResources {
						// curentGeneratedResourceWithResource
						// rename
						currentGeneratedResources[i] += ":" + arnType.Resource
					}
					// arn:aws:s3:us-east-1:123:bucket
					return currentGeneratedResources
				} else if arnType.Resource == "*" || arnType.Resource == "" {
					// resource = *
					ithResourceExpanded := []string{}
					// s3buckets := []string{"b1", "b2", "b3"}
					// s3paths := []string{"path1", "path2"}
					s3buckets := []string{}
					// query with where resource_name is "s3" or append the resource and path to exising
					// for bucket := range s3buckets {

					// 	for path := range s3paths {
					// 		ithResourceExpanded = append(ithResourceExpanded, currentGeneratedResources[i]+":"+s3buckets[bucket]+"/"+s3paths[path])
					// 	}

					// }

					// can add distinct here
					query := "select distinct id from resource where resource_type = 'awsS3'"
					// get bucket names
					s3buckets = fetchFromDb(s3buckets, query)

					for bucket := range s3buckets {

						ithResourceExpanded = append(ithResourceExpanded, currentGeneratedResources[i]+":"+s3buckets[bucket])

					}

					resourceFieldExpanded = append(resourceFieldExpanded, ithResourceExpanded...)

					// fmt.Println(ithResourceExpanded)
				} else if strings.ContainsAny(arnType.Resource, "*?") {

					ithResourceExpanded := []string{}
					for i := range currentGeneratedResources {

						arn := strings.SplitN(currentGeneratedResources[i], ":", 5)
						if arn != nil {
							switch arn[2] {
							case "s3":
								s3buckets := []string{}
								clause := replaceWildcards(arnType.Resource)

								query := "select distinct(id) from resource where resource_type = 'awsS3' and name like " + "'" + clause + "'"
								// get bucket names
								s3buckets = fetchFromDb(s3buckets, query)

								for bucket := range s3buckets {

									ithResourceExpanded = append(ithResourceExpanded, currentGeneratedResources[i]+":"+s3buckets[bucket])

								}

							}
						}
					}
					resourceFieldExpanded = append(resourceFieldExpanded, ithResourceExpanded...)
				}

			case "lambda":
				{
					ithResourceExpanded := []string{}
					var lambdaNames = make(map[string][]string)
					if arnType.Resource != "*" && arnType.Resource != "" && !strings.ContainsAny(arnType.Resource, "*?[]{}^") {

						for i := range currentGeneratedResources {
							currentGeneratedResources[i] += ":" + arnType.Resource
						}
						if strings.ContainsAny(arnType.ResourcePath, "?*") {
							resourcePath := replaceWildcards(arnType.ResourcePath)
							// resourcePath := strings.Replace(arnType.ResourcePath, "?", "_", -1)
							// resourcePath = strings.Replace(resourcePath, "*", "%", -1)
							clause := "and name like '" + resourcePath + "'"
							query := "select distinct name from resource where resource_type = 'awsLambda' and id like " + "'%" + arnType.Resource + "%' " + clause
							lambdaNames = getResourceTypes(lambdaNames, query, arnType.Resource)
							for i := range currentGeneratedResources {
								for name := range lambdaNames[arnType.Resource] {
									ithResourceExpanded = append(ithResourceExpanded, currentGeneratedResources[i]+":"+lambdaNames[arnType.Resource][name])
								}

							}
						} else if arnType.ResourcePath == "*" {
							query := "select distinct name from resource where resource_type = 'awsLambda' and id like " + "'%" + arnType.Resource + "%'"
							lambdaNames = getResourceTypes(lambdaNames, query, arnType.Resource)
							for i := range currentGeneratedResources {
								for name := range lambdaNames[arnType.Resource] {
									ithResourceExpanded = append(ithResourceExpanded, currentGeneratedResources[i]+":"+lambdaNames[arnType.Resource][name])
								}

							}
						} else {
							for i := range currentGeneratedResources {

								ithResourceExpanded = append(ithResourceExpanded, currentGeneratedResources[i]+":"+arnType.ResourcePath)

							}
						}
						resourceFieldExpanded = append(resourceFieldExpanded, ithResourceExpanded...)
						return resourceFieldExpanded
					} else if arnType.Resource == "*" || arnType.Resource == "" {

						var lambdaRes = []string{"function", "layer", "event-source-mapping"}
						for res := range lambdaRes {
							if len(lambdaNames[lambdaRes[res]]) == 0 {
								// can add distinct here
								query := "select distinct name from resource where resource_type = 'awsLambda' and id like " + "'%" + lambdaRes[res] + "%'"
								lambdaNames = getResourceTypes(lambdaNames, query, lambdaRes[res])
							}
						}
						for res := range lambdaRes {

							for name := range lambdaNames[lambdaRes[res]] {

								ithResourceExpanded = append(ithResourceExpanded, currentGeneratedResources[i]+":"+lambdaRes[res]+":"+lambdaNames[lambdaRes[res]][name])

							}
						}

						resourceFieldExpanded = append(resourceFieldExpanded, ithResourceExpanded...)
					}
				}

			case "dynamodb":
				// resource will always be table. just handling ResourcePaths
				{
					if arnType.ResourcePath != "*" && arnType.ResourcePath != "" && !strings.ContainsAny(arnType.ResourcePath, "?*[]{}^") {

						for i := range currentGeneratedResources {
							currentGeneratedResources[i] += ":" + "table/" + arnType.ResourcePath
						}
						return currentGeneratedResources
					} else if arnType.ResourcePath == "*" || arnType.ResourcePath == "" {
						ithResourceExpanded := []string{}
						dbTables := []string{}
						query := "select distinct name from resource where resource_type = 'awsDynamoDBTable'"
						// get table names
						dbTables = fetchFromDb(dbTables, query)
						for table := range dbTables {

							ithResourceExpanded = append(ithResourceExpanded, currentGeneratedResources[i]+":"+"table/"+dbTables[table])

						}

						resourceFieldExpanded = append(resourceFieldExpanded, ithResourceExpanded...)

					} else if strings.ContainsAny(arnType.ResourcePath, "?*") {
						ithResourceExpanded := []string{}
						dbTables := []string{}
						if len(dbTables) == 0 {
							clause := replaceWildcards(arnType.ResourcePath)
							// clause := strings.Replace(arnType.ResourcePath, "?", "_", -1)
							// clause = strings.Replace(clause, "*", "%", -1)
							query := "select distinct name from resource where resource_type = 'awsDynamoDBTable' and name like '" + clause + "'"
							// get table names
							dbTables = fetchFromDb(dbTables, query)
							for table := range dbTables {

								ithResourceExpanded = append(ithResourceExpanded, currentGeneratedResources[i]+":"+"table/"+dbTables[table])

							}

							resourceFieldExpanded = append(resourceFieldExpanded, ithResourceExpanded...)

						}
					}
				}
			case "cloudtrail":
				{
					if arnType.ResourcePath != "*" && arnType.ResourcePath != "" && !strings.ContainsAny(arnType.ResourcePath, "*?[]{}^") {

						for i := range currentGeneratedResources {
							currentGeneratedResources[i] += ":" + "trail/" + arnType.ResourcePath
						}
						return currentGeneratedResources
					} else if arnType.ResourcePath == "*" || arnType.ResourcePath == "" {
						ithResourceExpanded := []string{}
						trails := []string{}
						if len(trails) == 0 {
							query := "select distinct name from resource where resource_type = 'awsCloudtrail'"
							// get names
							trails = fetchFromDb(trails, query)
						}
						for trail := range trails {

							ithResourceExpanded = append(ithResourceExpanded, currentGeneratedResources[i]+":"+"trail/"+trails[trail])

						}

						resourceFieldExpanded = append(resourceFieldExpanded, ithResourceExpanded...)
					} else if strings.ContainsAny(arnType.ResourcePath, "?*") {
						ithResourceExpanded := []string{}
						trails := []string{}

						clause := replaceWildcards(arnType.ResourcePath)
						query := "select distinct name from resource where resource_type = 'awsCloudtrail' and name like '" + clause + "'"
						// get names
						trails = fetchFromDb(trails, query)

						for trail := range trails {

							ithResourceExpanded = append(ithResourceExpanded, currentGeneratedResources[i]+":"+"trail/"+trails[trail])

						}

						resourceFieldExpanded = append(resourceFieldExpanded, ithResourceExpanded...)
					}
				}
			case "redshift":
				{
					// currently uses {awsRedshiftCluster} as a filter in db to query out resources of cluster only,
					//  but code supports for any of the types of redshift instances like cluster, dbuster, dgroup etc,
					// just replace by appropriate filter
					ithResourceExpanded := []string{}
					var redshiftResourceTypes = make(map[string][]string)
					if arnType.Resource != "*" && arnType.Resource != "" && !strings.ContainsAny(arnType.Resource, "*?[]{}^") {

						for i := range currentGeneratedResources {
							currentGeneratedResources[i] += ":" + arnType.Resource
						}
						if strings.ContainsAny(arnType.ResourcePath, "?*") {
							resourcePath := replaceWildcards(arnType.ResourcePath)
							// resourcePath := strings.Replace(arnType.ResourcePath, "?", "_", -1)
							// resourcePath = strings.Replace(resourcePath, "*", "%", -1)
							clause := "and name like '" + resourcePath + "'"
							query := "select distinct name from resource where resource_type = 'awsRedshiftCluster' and id like " + "'%" + arnType.Resource + "%' " + clause
							redshiftResourceTypes = getResourceTypes(redshiftResourceTypes, query, arnType.Resource)
							for i := range currentGeneratedResources {
								for name := range redshiftResourceTypes[arnType.Resource] {
									ithResourceExpanded = append(ithResourceExpanded, currentGeneratedResources[i]+":"+redshiftResourceTypes[arnType.Resource][name])
								}

							}
						} else if arnType.ResourcePath == "*" || arnType.ResourcePath == "" {
							query := "select distinct name from resource where resource_type = 'awsRedshiftCluster' and id like " + "'%" + arnType.Resource + "%'"
							redshiftResourceTypes = getResourceTypes(redshiftResourceTypes, query, arnType.Resource)
							for i := range currentGeneratedResources {
								for name := range redshiftResourceTypes[arnType.Resource] {
									ithResourceExpanded = append(ithResourceExpanded, currentGeneratedResources[i]+":"+redshiftResourceTypes[arnType.Resource][name])
								}

							}
						} else {
							for i := range currentGeneratedResources {

								ithResourceExpanded = append(ithResourceExpanded, currentGeneratedResources[i]+":"+arnType.ResourcePath)

							}

						}
						resourceFieldExpanded = append(resourceFieldExpanded, ithResourceExpanded...)
						return resourceFieldExpanded
					} else if arnType.Resource == "*" || arnType.Resource == "" {

						var redShiftRes = []string{"cluster", "dbname", "dbuser", "dbgroup", "parametergroup", "securitygroup", "snapshot", "subnetgroup"}
						for res := range redShiftRes {
							if len(redshiftResourceTypes[redShiftRes[res]]) == 0 {
								// can add distinct here
								query := "select distinct name from resource where resource_type = 'awsRedshiftCluster' and id like " + "'%" + redShiftRes[res] + "%'"
								redshiftResourceTypes = getResourceTypes(redshiftResourceTypes, query, redShiftRes[res])
							}
						}
						for res := range redShiftRes {

							for name := range redshiftResourceTypes[redShiftRes[res]] {

								ithResourceExpanded = append(ithResourceExpanded, currentGeneratedResources[i]+":"+redShiftRes[res]+":"+redshiftResourceTypes[redShiftRes[res]][name])

							}
						}

						resourceFieldExpanded = append(resourceFieldExpanded, ithResourceExpanded...)
					}

				}
			default:

			}
		}

	}

	return resourceFieldExpanded
}

//  else if arnType.ResourcePath == "*" {
// 	for i := range currentGeneratedResources {
// 		arn, err := awsutil.ParseArn(currentGeneratedResources[i])
// 		if err == nil {
// 			// query with all fields i.e service, resource etc
// 			fmt.Println(arn)
// 		}

// 	}
// }

// Entry Function
func Expand(resources []string) {
	db = connectDatabase()

	ExpandedResourceSet := make(Set)
	for _, resource := range resources {

		currentGeneratedResources := []string{"arn"}
		// if resource == "*" {
		// 	currentGeneratedResources = handleStarWildcard(currentGeneratedResources, "Partition")
		// } else {

		// var arnType awsutil.Arn
		var arnType *awsutil.Arn
		var err error = nil
		if resource != "*" {
			arnType, err = awsutil.ParseArn(resource)
		} else {
			emptyArn := awsutil.Arn{
				Partition:    "",
				Service:      "",
				Region:       "",
				AccountId:    "",
				Resource:     "",
				ResourcePath: "",
			}
			arnType = &emptyArn
		}
		if err == nil {
			// fmt.Println(arnType)
			// fmt.Println("Partition: ", arnType.Partition, " Service: ", arnType.Service, "region: ", arnType.Region, "AccountId: ", arnType.AccountId, " Resource: ", arnType.Resource, "ResourcePath: ", arnType.ResourcePath)

			// handling Partition field
			if arnType.Partition == "*" || arnType.Partition == "" {
				currentGeneratedResources = handleStarWildcard(currentGeneratedResources, "Partition")
			} else if !strings.ContainsAny(arnType.Partition, "?[]{}^") {
				currentGeneratedResources[0] += ":" + arnType.Partition

			}
			// handling Service Field
			if arnType.Service == "*" || arnType.Service == "" {
				currentGeneratedResources = handleStarWildcard(currentGeneratedResources, "Service")

			} else if !strings.ContainsAny(arnType.Service, "?[]{}^") {
				for i := range currentGeneratedResources {
					currentGeneratedResources[i] += ":" + arnType.Service
				}
			}

			// handling Regions field
			if arnType.Region == "*" || arnType.Region == "" {
				currentGeneratedResources = handleStarWildcard(currentGeneratedResources, "Region")
			} else if !strings.ContainsAny(arnType.Partition, "?[]{}^") {
				for i := range currentGeneratedResources {
					currentGeneratedResources[i] += ":" + arnType.Region
				}

			}
			// handling Accountid field
			if arnType.AccountId == "*" || arnType.AccountId == "" {
				currentGeneratedResources = handleStarWildcard(currentGeneratedResources, "AccountId")
			} else if !strings.ContainsAny(arnType.Partition, "?[]{}^") {
				for i := range currentGeneratedResources {
					currentGeneratedResources[i] += ":" + arnType.AccountId
				}
			}

			// handling Resource Field
			currentGeneratedResources = handleResourcesField(arnType, currentGeneratedResources)
			// fmt.Println(currentGeneratedResources)
		} else {
			fmt.Println(err)
		}
		// }
		// fmt.Println(currentGeneratedResources)
		// sort.Strings((currentGeneratedResources))
		for _, element := range currentGeneratedResources {
			ExpandedResourceSet[element] = struct{}{}
		}
		fmt.Println(ExpandedResourceSet)

	}
	defer db.Close()

}

func connectDatabase() *sql.DB {
	// db, err := sql.Open("postgres", "postgres://username:password@host:port/database?sslmode=disable")
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		// Handle the error
		panic(err)
	}
	// defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected!")
	return db
}

const (
	host     = "localhost"
	port     = 5432
	user     = "stackidentity"
	password = "stackidentity"
	dbname   = "si_db"
)

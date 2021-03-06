#!/usr/bin/env bash

# more bash-friendly output for jq
JQ="jq --raw-output --exit-status"

configure_aws_cli(){
	aws --version
	aws configure set default.region ap-northeast-1
	aws configure set default.output json
}

deploy_cluster() {

    family="qiitan-api-task-family"

    make_task_def
    register_definition
    if [[ $(aws ecs update-service --cluster qiitan-api-cluster --service qiitan-api-service --task-definition $revision | \
                   $JQ '.service.taskDefinition') != $revision ]]; then
        echo "Error updating service."
        return 1
    fi

    # wait for older revisions to disappear
    # not really necessary, but nice for demos
    for attempt in {1..30}; do
        if stale=$(aws ecs describe-services --cluster qiitan-api-cluster --services qiitan-api-service | \
                       $JQ ".services[0].deployments | .[] | select(.taskDefinition != \"$revision\") | .taskDefinition"); then
            echo "Waiting for stale deployments:"
            echo "$stale"
            sleep 5
        else
            echo "Deployed!"
            return 0
        fi
    done
    echo "Service update took too long."
    return 1
}

make_task_def(){
	task_template='[
		{
			"name": "api",
			"image": "%s.dkr.ecr.ap-northeast-1.amazonaws.com/qiitan-api_api:%s",
			"essential": true,
			"memory": 200,
			"cpu": 10,
            "logConfiguration": {
                "logDriver": "awslogs",
                "options": {
                    "awslogs-group": "qiitan-api/api",
                    "awslogs-region": "ap-northeast-1"
                }
            },
            "command": ["go", "run", "main.go"]
		},
        {
			"name": "nginx",
			"image": "%s.dkr.ecr.ap-northeast-1.amazonaws.com/qiitan-api_nginx:%s",
			"essential": true,
			"memory": 200,
			"cpu": 10,
            "logConfiguration": {
                "logDriver": "awslogs",
                "options": {
                    "awslogs-group": "qiitan-api/nginx",
                    "awslogs-region": "ap-northeast-1"
                }
            },
			"portMappings": [
				{
					"containerPort": 80,
					"hostPort": 0
				}
			],
            "links": ["api:api"]
		},
        {
			"name": "mysql",
			"image": "%s.dkr.ecr.ap-northeast-1.amazonaws.com/qiitan-api_mysql:%s",
			"essential": true,
			"memory": 200,
			"cpu": 10,
            "portMappings": [
				{
					"containerPort": 3306,
					"hostPort": 3306
				}
			],
            "environment": [
                {"name": "MYSQL_ALLOW_EMPTY_PASSWORD", "value": "yes"},
                {"name": "MYSQL_USER", "value": "root"}
            ],
            "links": ["api:api"]
		},
        {
			"name": "aerospike",
			"image": "%s.dkr.ecr.ap-northeast-1.amazonaws.com/qiitan-api_aerospike:%s",
			"essential": true,
			"memory": 200,
			"cpu": 10,
            "portMappings": [
				{
					"containerPort": 3000,
					"hostPort": 3000
				}
			],
            "links": ["api:api"]
		}
	]'
	
	task_def=$(printf "$task_template" $AWS_ACCOUNT_ID $CIRCLE_SHA1 $AWS_ACCOUNT_ID $CIRCLE_SHA1 $AWS_ACCOUNT_ID $CIRCLE_SHA1 $AWS_ACCOUNT_ID $CIRCLE_SHA1)
}

push_ecr_image(){
	eval $(aws ecr get-login --region ap-northeast-1 --no-include-email)
	docker push $AWS_ACCOUNT_ID.dkr.ecr.ap-northeast-1.amazonaws.com/qiitan-api_api:$CIRCLE_SHA1
    docker push $AWS_ACCOUNT_ID.dkr.ecr.ap-northeast-1.amazonaws.com/qiitan-api_nginx:$CIRCLE_SHA1
    docker push $AWS_ACCOUNT_ID.dkr.ecr.ap-northeast-1.amazonaws.com/qiitan-api_mysql:$CIRCLE_SHA1
    docker push $AWS_ACCOUNT_ID.dkr.ecr.ap-northeast-1.amazonaws.com/qiitan-api_aerospike:$CIRCLE_SHA1
}

register_definition() {

    if revision=$(aws ecs register-task-definition --container-definitions "$task_def" --family $family | $JQ '.taskDefinition.taskDefinitionArn'); then
        echo "Revision: $revision"
    else
        echo "Failed to register task definition"
        return 1
    fi

}

configure_aws_cli
push_ecr_image
deploy_cluster

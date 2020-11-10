import * as cdk from '@aws-cdk/core';
import assets = require("@aws-cdk/aws-s3-assets");
import events = require("@aws-cdk/aws-events");
import events_targets = require("@aws-cdk/aws-events-targets");
import iam = require("@aws-cdk/aws-iam");
import lambda = require("@aws-cdk/aws-lambda");
import path = require("path");

export class TonkatsuStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const giveCatTreatsAsset = new assets.Asset(this, "giveCatTreatsZip", {
      path: path.join(__dirname, "../../build/"),
    })

    const giveCatTreatsLambda = new lambda.Function(this, "giveCatTreatsLambda", {
      code: lambda.Code.fromBucket(
        giveCatTreatsAsset.bucket,
        giveCatTreatsAsset.s3ObjectKey,
      ),
      runtime: lambda.Runtime.GO_1_X,
      handler: "give-cat-treats",
      timeout: cdk.Duration.seconds(45),
    })

    giveCatTreatsLambda.addToRolePolicy(new iam.PolicyStatement({
      actions: [
        "cloudwatch:PutMetricData",
      ],
      resources: ["*"],
    }))

    new events.Rule(this, "giveCatTreatsSchedule", {
      schedule: events.Schedule.cron({minute: '5'}),
      targets: [new events_targets.LambdaFunction(giveCatTreatsLambda)],
    })
  }
}

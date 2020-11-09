import * as cdk from '@aws-cdk/core';
import assets = require("@aws-cdk/aws-s3-assets");
import lambda = require("@aws-cdk/aws-lambda");
import path = require("path");

export class InfraStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const giveCatTreatsAsset = new assets.Asset(this, "giveCatTreatsZip", {
      path: path.join(process.env["BINARIES_DIR"]!, "/build/give-cat-treats"),
    })

    const giveCatTreatsLambda = new lambda.Function(this, "giveCatTreatsLambda", {
      code: lambda.Code.fromBucket(
        giveCatTreatsAsset.bucket,
        giveCatTreatsAsset.s3ObjectKey,
      ),
      runtime: lambda.Runtime.GO_1_X,
      handler: "main",
      timeout: cdk.Duration.seconds(45),
    })
  }
}

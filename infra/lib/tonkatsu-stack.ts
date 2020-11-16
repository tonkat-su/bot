import * as cdk from '@aws-cdk/core';
import assets = require("@aws-cdk/aws-s3-assets");
import events = require("@aws-cdk/aws-events");
import events_targets = require("@aws-cdk/aws-events-targets");
import iam = require("@aws-cdk/aws-iam");
import lambda = require("@aws-cdk/aws-lambda");
import logs = require("@aws-cdk/aws-logs");
import path = require("path");
import route53 = require("@aws-cdk/aws-route53");
import secretsManager = require("@aws-cdk/aws-secretsmanager");
import { Duration } from '@aws-cdk/core';

export class TonkatsuStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);
    
    const tonkatsuZone = route53.HostedZone.fromHostedZoneAttributes(this, 'tonkatsuZone', {
      hostedZoneId: 'ZVAMW53PNR70P',
      zoneName: 'tonkat.su',
    })

    new route53.ARecord(this, 'botRecord', {
      zone: tonkatsuZone,
      recordName: "bot",
      target: route53.RecordTarget.fromIpAddresses("45.33.41.248"),
    })

    const lambdasAsset = new assets.Asset(this, "lambdasZip", {
      path: path.join(__dirname, "../../build/"),
    })

    const giveCatTreatsLambda = new lambda.Function(this, "giveCatTreatsLambda", {
      code: lambda.Code.fromBucket(
        lambdasAsset.bucket,
        lambdasAsset.s3ObjectKey,
      ),
      runtime: lambda.Runtime.GO_1_X,
      handler: "give-cat-treats",
      timeout: cdk.Duration.seconds(45),
      environment: {
        "MINECRAFT_SERVER_NAME": "NewPumpcraft",
        "MINECRAFT_SERVER_HOST": "mc.sep.gg",
      },
      logRetention: logs.RetentionDays.THREE_DAYS,
    })

    giveCatTreatsLambda.addToRolePolicy(new iam.PolicyStatement({
      actions: [
        "cloudwatch:PutMetricData",
      ],
      resources: ["*"],
    }))

    const discordToken = secretsManager.Secret.fromSecretCompleteArn(this, "discordToken", "arn:aws:secretsmanager:us-west-2:635281304921:secret:tonkatsu/bot/prod/discordToken-SGeL0l")

    const refreshWhosOnlineLambda = new lambda.Function(this, "refreshWhosOnlineLambda", {
      code: lambda.Code.fromBucket(
        lambdasAsset.bucket,
        lambdasAsset.s3ObjectKey,
      ),
      runtime: lambda.Runtime.GO_1_X,
      handler: "refresh-whos-online",
      timeout: cdk.Duration.seconds(30),
      environment: {
        "MINECRAFT_SERVER_HOST": "mc.sep.gg",
        "MINECRAFT_SERVER_NAME": "NewPumpcraft",
        "DISCORD_TOKEN_SECRET_ARN": discordToken.secretArn,
      },
      logRetention: logs.RetentionDays.THREE_DAYS,
    })

    refreshWhosOnlineLambda.addToRolePolicy(new iam.PolicyStatement({
      actions: [
        "secretsmanager:GetResourcePolicy",
        "secretsmanager:GetSecretValue",
        "secretsmanager:DescribeSecret",
        "secretsmanager:ListSecretVersionIds"
      ],
      resources: [discordToken.secretArn],
    }))

    new events.Rule(this, "everyFiveMinutes", {
      schedule: events.Schedule.rate(Duration.minutes(5)),
      targets: [
        new events_targets.LambdaFunction(giveCatTreatsLambda),
      ],
    })
  }
}

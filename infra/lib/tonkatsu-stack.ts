import * as cdk from 'aws-cdk-lib';
import { App, Stack } from 'aws-cdk-lib';
import { aws_route53 as route53 } from 'aws-cdk-lib'
import { aws_secretsmanager as secretsManager } from 'aws-cdk-lib';
import { aws_lambda as lambda } from 'aws-cdk-lib';
import { aws_iam as iam } from 'aws-cdk-lib';
import { aws_events as events } from 'aws-cdk-lib';
import { aws_events_targets as events_targets } from 'aws-cdk-lib';
import { aws_dynamodb as dynamodb } from 'aws-cdk-lib';
import { aws_logs as logs } from 'aws-cdk-lib';
import { aws_s3_assets as assets } from 'aws-cdk-lib';
import { aws_certificatemanager as certificatemanager } from 'aws-cdk-lib';
import { aws_apigateway as apigateway } from 'aws-cdk-lib';
import { aws_route53_targets as targets } from 'aws-cdk-lib';
import * as path from 'path';

export class TonkatsuStack extends Stack {
  constructor(scope: App, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const botGroup = new iam.Group(this, 'tonkatsuBotGroup', {})

    botGroup.addToPolicy(new iam.PolicyStatement({
      actions: [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "cloudwatch:PutMetricData",
        "cloudwatch:ListMetrics",
        "cloudwatch:GetMetricData",
      ],
      resources: ["*"],
    }))

    const botUser = new iam.User(this, 'tonkatsuBotUser', {
      groups: [botGroup],
    })

    const usersTable = new dynamodb.Table(this, 'users', {
      partitionKey: {
        name: 'DiscordUserId',
        type: dynamodb.AttributeType.STRING,
      }
    })

    usersTable.addGlobalSecondaryIndex({
      indexName: 'MinecraftIdIndex',
      partitionKey: {
        name: 'MinecraftUserId',
        type: dynamodb.AttributeType.STRING,
      }
    })
    usersTable.grantReadWriteData(botGroup)

    const tonkatsuZone = route53.HostedZone.fromHostedZoneAttributes(this, 'tonkatsuZone', {
      hostedZoneId: 'ZVAMW53PNR70P',
      zoneName: 'tonkat.su',
    })

    new route53.ARecord(this, 'botRecord', {
      zone: tonkatsuZone,
      recordName: "bot",
      target: route53.RecordTarget.fromIpAddresses("45.79.73.44"),
    })

    new route53.AaaaRecord(this, 'botAAAARecord', {
      zone: tonkatsuZone,
      recordName: "bot",
      target: route53.RecordTarget.fromIpAddresses("2600:3c01::f03c:93ff:fe7a:5fd8"),
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
        "MINECRAFT_SERVER_NAME": "frogland",
        "MINECRAFT_SERVER_HOST": "mc.froggyfren.com",
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
        "MINECRAFT_SERVER_HOST": "mc.froggyfren.com",
        "MINECRAFT_SERVER_NAME": "frogland",
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
      schedule: events.Schedule.rate(cdk.Duration.minutes(5)),
      targets: [
        new events_targets.LambdaFunction(giveCatTreatsLambda),
      ],
    })

    const smpRconPassword = secretsManager.Secret.fromSecretCompleteArn(this, "smpRconPassword", "arn:aws:secretsmanager:us-west-2:635281304921:secret:prod/mc.tonkat.su/rconpassword-dEsrPy")

    const interactionsWhitelistLambda = new lambda.Function(this, 'interactionsWhitelistLambda', {
      code: lambda.Code.fromBucket(
        lambdasAsset.bucket,
        lambdasAsset.s3ObjectKey,
      ),
      runtime: lambda.Runtime.GO_1_X,
      handler: "smp-whitelist",
      timeout: cdk.Duration.seconds(15),
      environment: {
        "MINECRAFT_SERVER_RCON_ADDRESS": "mc.tonkat.su:9763",
        "DISCORD_APPLICATION_PUBKEY": "14f8daad94d0146557e27c172f597d5707c91025774ac6bc99fb0caffd21fd7c",
        "RCON_PASSWORD_SECRET_ARN": smpRconPassword.secretArn,
      },
      logRetention: logs.RetentionDays.THREE_DAYS,
    })

    interactionsWhitelistLambda.addToRolePolicy(new iam.PolicyStatement({
      actions: [
        "secretsmanager:GetResourcePolicy",
        "secretsmanager:GetSecretValue",
        "secretsmanager:DescribeSecret",
        "secretsmanager:ListSecretVersionIds"
      ],
      resources: [smpRconPassword.secretArn],
    }))

    const interactionsCert = new certificatemanager.DnsValidatedCertificate(this, 'interactionsCert', {
      domainName: 'interactions.tonkat.su',
      hostedZone: tonkatsuZone,
      region: 'us-west-2',
    })

    const interactionsApi = new apigateway.LambdaRestApi(this, 'interactionsApi', {
      domainName: {
        domainName: 'interactions.tonkat.su',
        certificate: interactionsCert,
      },
      handler: interactionsWhitelistLambda,
      proxy: false,
    })

    const whitelistApiResource = interactionsApi.root.addResource('whitelist')
    whitelistApiResource.addMethod('POST')

    new route53.ARecord(this, 'interactionsWhitelistAliasRecord', {
      zone: tonkatsuZone,
      recordName: 'interactions',
      target: route53.RecordTarget.fromAlias(new targets.ApiGateway(interactionsApi))
    })
  }
}

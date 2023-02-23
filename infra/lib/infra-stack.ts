import * as cdk from "aws-cdk-lib";
import { App, Stack } from "aws-cdk-lib";
import * as acm from 'aws-cdk-lib/aws-certificatemanager';
import { aws_route53 as route53 } from "aws-cdk-lib";
import { aws_secretsmanager as secretsManager } from "aws-cdk-lib";
import { aws_lambda as lambda } from "aws-cdk-lib";
import { aws_iam as iam } from "aws-cdk-lib";
import { aws_events as events } from "aws-cdk-lib";
import { aws_events_targets as events_targets } from "aws-cdk-lib";
import { HttpLambdaIntegration } from '@aws-cdk/aws-apigatewayv2-integrations-alpha';
import  * as apigwv2 from "@aws-cdk/aws-apigatewayv2-alpha";
import { aws_logs as logs } from "aws-cdk-lib";
import { aws_s3_assets as assets } from "aws-cdk-lib";
import { aws_route53_targets as targets } from "aws-cdk-lib";
import * as path from "path";

export class InfraStack extends Stack {
  constructor(scope: App, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const botGroup = new iam.Group(this, "infraBotGroup", {});

    botGroup.addToPolicy(
      new iam.PolicyStatement({
        actions: [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents",
          "cloudwatch:PutMetricData",
          "cloudwatch:ListMetrics",
          "cloudwatch:GetMetricData",
        ],
        resources: ["*"],
      })
    );

    const botUser = new iam.User(this, "infraBotUser", {
      groups: [botGroup],
    });

    const infraZone = route53.HostedZone.fromHostedZoneAttributes(
      this,
      "infraZone",
      {
        hostedZoneId: "Z3CDOBSYLTO062",
        zoneName: "sjchen.com",
      }
    );

    const lambdasAsset = new assets.Asset(this, "lambdasZip", {
      path: path.join(__dirname, "../../build/"),
    });

    const giveCatTreatsLambda = new lambda.Function(
      this,
      "giveCatTreatsLambda",
      {
        code: lambda.Code.fromBucket(
          lambdasAsset.bucket,
          lambdasAsset.s3ObjectKey
        ),
        runtime: lambda.Runtime.GO_1_X,
        handler: "give-cat-treats",
        timeout: cdk.Duration.seconds(45),
        environment: {
          MINECRAFT_SERVER_NAME: "frogland",
          MINECRAFT_SERVER_HOST: "mc.froggyfren.com",
        },
        logRetention: logs.RetentionDays.THREE_DAYS,
      }
    );

    giveCatTreatsLambda.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ["cloudwatch:PutMetricData"],
        resources: ["*"],
      })
    );

    const discordToken = secretsManager.Secret.fromSecretCompleteArn(
      this,
      "discordToken",
      "arn:aws:secretsmanager:us-west-2:635281304921:secret:tonkatsu/bot/prod/discordToken-SGeL0l"
    );

    const refreshWhosOnlineLambda = new lambda.Function(
      this,
      "refreshWhosOnlineLambda",
      {
        code: lambda.Code.fromBucket(
          lambdasAsset.bucket,
          lambdasAsset.s3ObjectKey
        ),
        runtime: lambda.Runtime.GO_1_X,
        handler: "refresh-whos-online",
        timeout: cdk.Duration.seconds(30),
        environment: {
          MINECRAFT_SERVER_HOST: "mc.froggyfren.com",
          MINECRAFT_SERVER_NAME: "frogland",
          DISCORD_TOKEN_SECRET_ARN: discordToken.secretArn,
        },
        logRetention: logs.RetentionDays.THREE_DAYS,
      }
    );

    refreshWhosOnlineLambda.addToRolePolicy(
      new iam.PolicyStatement({
        actions: [
          "secretsmanager:GetResourcePolicy",
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret",
          "secretsmanager:ListSecretVersionIds",
        ],
        resources: [discordToken.secretArn],
      })
    );

    new events.Rule(this, "everyFiveMinutes", {
      schedule: events.Schedule.rate(cdk.Duration.minutes(5)),
      targets: [new events_targets.LambdaFunction(giveCatTreatsLambda)],
    });

    const smpRconPassword = secretsManager.Secret.fromSecretCompleteArn(
      this,
      "smpRconPassword",
      "arn:aws:secretsmanager:us-west-2:635281304921:secret:prod/mc.tonkat.su/rconpassword-dEsrPy"
    );

    const interactionsWhitelistLambda = new lambda.Function(
      this,
      "interactionsWhitelistLambda",
      {
        code: lambda.Code.fromBucket(
          lambdasAsset.bucket,
          lambdasAsset.s3ObjectKey
        ),
        runtime: lambda.Runtime.GO_1_X,
        handler: "smp-whitelist",
        timeout: cdk.Duration.seconds(15),
        environment: {
          MINECRAFT_SERVER_RCON_ADDRESS: "mc.tonkat.su:9763",
          DISCORD_APPLICATION_PUBKEY:
            "14f8daad94d0146557e27c172f597d5707c91025774ac6bc99fb0caffd21fd7c",
          RCON_PASSWORD_SECRET_ARN: smpRconPassword.secretArn,
        },
        logRetention: logs.RetentionDays.THREE_DAYS,
      }
    );

    interactionsWhitelistLambda.addToRolePolicy(
      new iam.PolicyStatement({
        actions: [
          "secretsmanager:GetResourcePolicy",
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret",
          "secretsmanager:ListSecretVersionIds",
        ],
        resources: [smpRconPassword.secretArn],
      })
    );

    const interactionsCert = new acm.Certificate(
      this,
      "interactionsCert",
      {
        domainName: "interactions.sjchen.com",
        validation: acm.CertificateValidation.fromDns(infraZone),
      }
    );

    const interactionsApiLambda = new lambda.Function(
      this,
      "interactionsApiLambda",
      {
        code: lambda.Code.fromBucket(
          lambdasAsset.bucket,
          lambdasAsset.s3ObjectKey
        ),
        runtime: lambda.Runtime.GO_1_X,
        handler: "interactions",
        timeout: cdk.Duration.seconds(15),
        environment: {
          DISCORD_APPLICATION_PUBKEY:
            "14f8daad94d0146557e27c172f597d5707c91025774ac6bc99fb0caffd21fd7c",
        },
        logRetention: logs.RetentionDays.THREE_DAYS,
      }
    );

    const interactionsLambdaApi = new HttpLambdaIntegration(
      "Interactions",
      interactionsApiLambda
    );

    const dn = new apigwv2.DomainName(this, 'DN', {
      domainName: 'interactions.sjchen.com',
      certificate: acm.Certificate.fromCertificateArn(this, 'cert', interactionsCert.certificateArn),
    });

    const httpApi = new apigwv2.HttpApi(this, 'DiscordInteractionsApiGateway', {
      defaultDomainMapping: {
        domainName: dn,
      },
    });
    httpApi.addRoutes({
      path: '/',
      methods: [ apigwv2.HttpMethod.POST ],
      integration: interactionsLambdaApi,
    });

    new route53.ARecord(this, "interactionsWhitelistAliasRecord", {
      zone: infraZone,
      recordName: "interactions",
      target: route53.RecordTarget.fromAlias(
        new targets.ApiGatewayv2DomainProperties(dn.regionalDomainName, dn.regionalHostedZoneId)
      ),
    });
  }
}

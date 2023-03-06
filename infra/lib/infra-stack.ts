import * as cdk from "aws-cdk-lib";
import { App, Stack } from "aws-cdk-lib";
import * as acm from "aws-cdk-lib/aws-certificatemanager";
import { aws_route53 as route53 } from "aws-cdk-lib";
import { aws_secretsmanager as secretsManager } from "aws-cdk-lib";
import { aws_lambda as lambda } from "aws-cdk-lib";
import { aws_iam as iam } from "aws-cdk-lib";
import { aws_events as events } from "aws-cdk-lib";
import { aws_events_targets as events_targets } from "aws-cdk-lib";
import * as apigwv2integration from "@aws-cdk/aws-apigatewayv2-integrations-alpha";
import * as apigwv2 from "@aws-cdk/aws-apigatewayv2-alpha";
import { aws_logs as logs } from "aws-cdk-lib";
import { aws_s3_assets as assets } from "aws-cdk-lib";
import { aws_route53_targets as targets } from "aws-cdk-lib";
import * as path from "path";

export class InfraStack extends Stack {
  constructor(scope: App, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const discordApplicationPubkey =
      "14f8daad94d0146557e27c172f597d5707c91025774ac6bc99fb0caffd21fd7c";

    const infraZone = route53.HostedZone.fromHostedZoneAttributes(
      this,
      "infraZone",
      {
        hostedZoneId: "Z05432021TEFY2IZX7IFD",
        zoneName: "bsdlp.dev",
      }
    );

    const lambdasAsset = new assets.Asset(this, "lambdasZip", {
      path: path.join(__dirname, "../../build/"),
    });

    const interactionsSecrets = secretsManager.Secret.fromSecretCompleteArn(
      this,
      "interactionsSecrets",
      "arn:aws:secretsmanager:us-west-2:635281304921:secret:discord-interactions-api-enWlPw"
    );

    const cdkDeployRole = iam.Role.fromRoleArn(
      this,
      "deployIamUser",
      "arn:aws:iam::635281304921:role/github-cdk-deploy"
    );
    interactionsSecrets.grantRead(cdkDeployRole);

    // lambda to grant cat treats
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
          MINECRAFT_SERVER_NAME: "froggyland",
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

    new events.Rule(this, "everyFiveMinutes", {
      schedule: events.Schedule.rate(cdk.Duration.minutes(5)),
      targets: [new events_targets.LambdaFunction(giveCatTreatsLambda)],
    });

    const interactionsCert = new acm.Certificate(this, "interactionsCert", {
      domainName: "interactions.bsdlp.dev",
      validation: acm.CertificateValidation.fromDns(infraZone),
    });

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
          DISCORD_WEBHOOK_PUBKEY: discordApplicationPubkey,
          DISCORD_TOKEN: interactionsSecrets
            .secretValueFromJson("DISCORD_TOKEN")
            .toString(),
          IMGUR_CLIENT_ID: interactionsSecrets
            .secretValueFromJson("IMGUR_CLIENT_ID")
            .toString(),
          RCON_PASSWORD: interactionsSecrets
            .secretValueFromJson("RCON_PASSWORD")
            .toString(),
          RCON_HOSTPORT: "mc.froggyfren.com:25575",
          MINECRAFT_SERVER_NAME: "froggyland",
          MINECRAFT_SERVER_HOST: "mc.froggyfren.com",
          DISCORD_GUILD_ID: "764720442250100757",
        },
        logRetention: logs.RetentionDays.THREE_DAYS,
      }
    );

    interactionsApiLambda.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ["cloudwatch:ListMetrics", "cloudwatch:GetMetricData"],
        resources: ["*"],
      })
    );

    const interactionsLambdaApi = new apigwv2integration.HttpLambdaIntegration(
      "Interactions",
      interactionsApiLambda,
      {
        payloadFormatVersion: apigwv2.PayloadFormatVersion.VERSION_1_0,
      }
    );

    const dn = new apigwv2.DomainName(this, "DN", {
      domainName: "interactions.bsdlp.dev",
      certificate: acm.Certificate.fromCertificateArn(
        this,
        "cert",
        interactionsCert.certificateArn
      ),
    });

    const httpApi = new apigwv2.HttpApi(this, "DiscordInteractionsApiGateway", {
      defaultDomainMapping: {
        domainName: dn,
      },
    });
    httpApi.addRoutes({
      path: "/interactions",
      methods: [apigwv2.HttpMethod.POST],
      integration: interactionsLambdaApi,
    });

    new route53.ARecord(this, "interactionsAliasRecord", {
      zone: infraZone,
      recordName: "interactions.bsdlp.dev.",
      target: route53.RecordTarget.fromAlias(
        new targets.ApiGatewayv2DomainProperties(
          dn.regionalDomainName,
          dn.regionalHostedZoneId
        )
      ),
    });
  }
}

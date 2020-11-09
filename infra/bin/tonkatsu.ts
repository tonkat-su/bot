#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import { TonkatsuStack } from '../lib/tonkatsu-stack';

const app = new cdk.App();
new TonkatsuStack(app, 'TonkatsuStack');

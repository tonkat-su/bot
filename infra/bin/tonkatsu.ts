#!/usr/bin/env node
import 'source-map-support/register';
import { App } from 'aws-cdk-lib';
import { TonkatsuStack } from '../lib/tonkatsu-stack';

const app = new App();
new TonkatsuStack(app, 'TonkatsuStack');

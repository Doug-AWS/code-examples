import { expect as expectCDK, matchTemplate, MatchStyle } from '@aws-cdk/assert';
import * as cdk from '@aws-cdk/core';
import * as ImageRecog from '../lib/image-recog-stack';

test('Empty Stack', () => {
  const app = new cdk.App();
  // WHEN
  const stack = new ImageRecog.ImageRecogStack(app, 'MyTestStack');
  // THEN
  expectCDK(stack).to(matchTemplate({
    "Resources": {}
  }, MatchStyle.EXACT))
});

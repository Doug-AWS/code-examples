import { expect as expectCDK, matchTemplate, MatchStyle } from '@aws-cdk/assert';
import * as cdk from '@aws-cdk/core';
import * as CreateUserPool from '../lib/create_user_pool-stack';

test('Empty Stack', () => {
    const app = new cdk.App();
    // WHEN
    const stack = new CreateUserPool.CreateUserPoolStack(app, 'MyTestStack');
    // THEN
    expectCDK(stack).to(matchTemplate({
      "Resources": {}
    }, MatchStyle.EXACT))
});

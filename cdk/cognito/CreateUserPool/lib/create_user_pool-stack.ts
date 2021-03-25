import * as cdk from '@aws-cdk/core';
import { UserPool } from '@aws-cdk/aws-cognito'
import { CfnOutput } from '@aws-cdk/core';

export class CreateUserPoolStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // Cognito User Pool with Email Sign-in Type.
    const userPool = new UserPool(this, 'userPool', {
      signInAliases: {
        email: true
      }
    })

    new CfnOutput(this, 'User pool name: ', { value: userPool.userPoolProviderName });
    new CfnOutput(this, 'User pool ID:   ', { value: userPool.userPoolId });
    new CfnOutput(this, 'User pool ARN:  ', { value: userPool.userPoolArn });
  }
}

import * as cdk from '@aws-cdk/core';
import * as codebuild from '@aws-cdk/aws-codebuild';
import * as amplify from '@aws-cdk/aws-amplify';


export class TrailAppStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const amplifyApp = new amplify.App(this, 'MyApp', {
      sourceCodeProvider: new amplify.GitHubSourceCodeProvider({
        owner: '<user>',
        repository: '<repo>',
        oauthToken: cdk.SecretValue.secretsManager('my-github-token')
      }),
      buildSpec: codebuild.BuildSpec.fromObject({ // Alternatively add a `amplify.yml` to the repo
        version: '1.0',
        frontend: {
          phases: {
            preBuild: {
              commands: [
                'yarn'
              ]
            },
            build: {
              commands: [
                'yarn build'
              ]
            }
          },
          artifacts: {
            baseDirectory: 'public',
            files: '**/*'
          }
        }
      })
    });
  }
}

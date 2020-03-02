# slack-webhook-interceptor

This is a [Tekton](https://github.com/tektoncd/triggers) [EventListener interceptor](https://github.com/tektoncd/triggers/blob/master/examples/eventlisteners/eventlistener-interceptor.yaml) that can turnaround Slack "slash command" requests into JSON for further processing.

## Blogpost

See this [https://bigkevmcd.github.io/tekton/triggers/slack/commands/2020/03/02/slack-slash-tekton.html](https://bigkevmcd.github.io/tekton/triggers/slack/commands/2020/03/02/slack-slash-tekton.html) for detailed instructions.

## Usage

See this [example](https://github.com/bigkevmcd/slack-webhook-interceptor/blob/master/example/listener-interceptor.yaml) EventListener for how to use this, it shows an EventListener that uses the interceptor to drive a TaskRun that outputs the command and details.

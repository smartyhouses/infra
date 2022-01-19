# Devbook SDK
**Devbook makes your dev docs interactive with just 3 lines of code.**

Devbook is a JS library that allows visitors of your docs to interact with and execute any code snippet or shell command in a private VM.

## How Devbook works
Every time a user visits a page where you use Devbook (like your docs), we quickly spin up a private VM just for that user.
They can experiment and explore your API/SDK right from your docs. Zero setup and overhead.

**Check this [Twitter thread](https://twitter.com/mlejva/status/1482767780265050126) with a video to see Devbook in action.**

## Installation
```sh
npm install @devbookhq/sdk
```
## Usage

### React
```tsx
// 1. Import the hook.
import {
  useDevbook,
  Env,
} from '@devbookhq/sdk'

// 2. Define your code.
const code = `
 > Code that you want to execute in a VM goes here.
`

function InteractiveCodeSnippet() {
  const { stdout, stderr, runCode } = useDevbook({ env: Env.NodeJS })

  return (
    <div>
      <button onClick={() => runCode(code)}>Run</button>
      <h3>Output</h3>
      {stdout.map((o, idx) => <span key={`out_${idx}`}>{o}</span>)}
      {stderr.map((e, idx) => <span key={`err_${idx}`}>{e}</span>)}
    </div>
  )
}

export default InteractiveCodeSnippet
```

### Vanilla JS
```ts
  import { Devbook, Env } from '@devbookhq/sdk'

  // 2. Define your code.
  const code = `
   > Code that you want to execute in a VM goes here.
  `

  // 3. Define callbacks.
  function handleStdout(out: string) {
    console.log('stdout', { err })
  }

  function handleStderr(err: string) {
    console.log('stderr', { err })
  }

  // 4. Create new Devbook instance.
  const dbk = new Devbook({ env: Env.NodeJS, onStdout: handleStdout, onStderr: handleStderr })
  dbk.runCode(code)
```

## Supported runtimes
- NodeJS
- *(coming soon)* Custom container based environment

## Usage of Devbook in example apps
- [React](examples/react-app)
- [MDX (Docusaurus and other docs themes)](examples/docusaurus)

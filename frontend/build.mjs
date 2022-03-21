import { pnpPlugin } from '@yarnpkg/esbuild-plugin-pnp'
import { build, analyzeMetafile } from 'esbuild'
import sveltePlugin from 'esbuild-svelte'
import sveltePreprocess from 'svelte-preprocess'
import fs from 'fs'
import path from 'path'
import { exec } from 'child_process'

const args = process.argv.slice(2)
const srcDir = './src/'
const outDir = './dist/'
const staticDir = './static/'

let lastError = ""
const handleError = (errors) => {
  const errStr = JSON.stringify(errors, null, 2)
  if (lastError === errStr) {
    console.log('Error is the same as last time')
    return
  }
  lastError = errStr
  // This is just stupid, really, but I wanted to do this sort of thing without vite or any other stuff.
  const html = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png" />
    <link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png" />
    <link rel="icon" type="image/png" sizes="16x16" href="/favicon-16x16.png" />
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.4.0/styles/dark.min.css">
    <link rel="manifest" href="/site.webmanifest" />
    <link rel="mask-icon" href="/safari-pinned-tab.svg" color="#d72312" />
    <meta name="msapplication-TileColor" content="#d72312" />
    <meta name="theme-color" content="#ffffff" />
    <title>Skiver</title>
    <link rel="stylesheet" href="./entry.css" />
  </head>
  <body>
   ${errors.errors.map(err => {
    const { lineText, file, line, suggestion } = err.location
    return `<div style="color: hotpink">
       <h3>${file}:${line}</h3>
       <p>${err.text}</p>
       ${!suggestion ? "" : `<p>{suggestion}</p>`}
       <pre class="language-plaintext"><code style="background: #303030">${lineText}</code></pre>
       <pre class="language-json"><code style="background: #303030">${errStr}</code></pre>
     </div>`
  })}
  <script type="module"defer>
    import hljs from 'https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.4.0/build/es/highlight.min.js';
    hljs.highlightAll();

const wsUrl = window.location.protocol.replace("http", "ws") +  "//" + window.location.host + "/ws/"
  console.log(wsUrl)
 const wsSubscribe = () => {
  try {
    const conn = new WebSocket(wsUrl)
    conn.onerror = function(evt) {
      console.error('[ws] connection error: ', evt)
      setTimeout(() => wsSubscribe(options), 100)
    }
    conn.onclose = function(evt) {
      console.debug('[ws]: connection closed', evt)
      wsDisconnects++
      setTimeout(() => wsSubscribe(options), 100)
    }
    conn.onmessage = function(evt) {
      window.location.reload()
    }
  } catch (err) {
    console.error('Failed in wsSubscribe ', err)
    setTimeout(() => wsSubscribe(options), 100)
  }
}
    wsSubscribe()
  </script>
  </body>
</html>

`
  const p = path.join(outDir, "index.html")
  fs.writeFileSync(p, html)
}

const execP = (args) => {

  console.debug('executing: ', args)
  return new Promise((res) =>
    exec(args, (err, out, stdErr) =>
      res([out, (err || stdErr) && { err, stdErr }])
    )
  )
}

const isDev =
  args.includes('dev') || args.includes('-d') || args.includes('development')
const withDTS =
  args.includes('dts') || args.includes('-t') || args.includes('types')

if (!fs.existsSync(outDir)) {
  fs.mkdirSync(outDir)
}

const createTypescriptApiDefinitions = async () => {
  if (!withDTS) {
    return
  }
  console.time('🌱 creating typescript api defintions...')
  const [res, err] = await execP('yarn gen')
  if (err) {
    console.error('🔥 Failed to create typescript-defintitions for api: ', err)
  } else {
    console.info('🪴 Created typescript-defintions for api', out)
  }
  console.timeEnd('🌱 creating typescript api defintions...')
}
const typecheck = async () => {
  console.time('🦴 typechecking')
  const [res, err] = await execP('yarn tsc --noEmit')
  if (res) {
    console.info(res)
  }
  if (err && !(res || '').includes('error')) {
    console.error('🔥🦴', err)
  }
  console.timeEnd('🦴 typechecking')
}
async function run() {

  createTypescriptApiDefinitions()
  typecheck()

  const result = await build({
    plugins: [
      pnpPlugin(),
      sveltePlugin({
        preprocess: sveltePreprocess(),
      }),
    ],
    entryPoints: [srcDir + 'entry.ts'],
    bundle: true,
    splitting: true,
    format: 'esm',
    outdir: outDir,
    logLevel: 'info',
    // sourcemap: 'external',
    legalComments: 'external',
    minify: true,
    metafile: true,
    ...(isDev && {
      metafile: false,
      watch: {
        onRebuild: (error, result) => {
          if (error) {
            console.error('!!! watch build failed:', error)
            handleError(error)
          } else {
            fs.copyFile(srcDir + 'index.html', outDir + '/index.html', (err) => {
              if (err) throw err
            })
            const reduced = Object.entries(result).reduce((r, [k, v]) => {
              if (!v) {
                return r
              }
              if (typeof v === 'function') {
                return r
              }
              if (Array.isArray(v) && !v.length) {
                return r
              }
              if (k === 'metafile') {
                return r
              }
              r[k] = v
              return r
            }, {})
            if (Object.keys(reduced).length) {
              console.info('🎉 watch build succeeded with result:', reduced)
            } else {
              console.info('🎉 watch build succeeded')
            }
          }
          createTypescriptApiDefinitions()
          typecheck()
        },
      },
      legalComments: 'none',
      minify: false,
      sourcemap: 'inline',
    }),
  }).catch(
    (error) => {
      console.log('errryryryryryr')
      handleError(error)
      if (isDev) {
        setTimeout(run, 1000)
        return
      }
      throw error
    }
  )
  if (!result) {
    return
  }

  if (result.metafile) {
    const analysis = await analyzeMetafile(result.metafile, { verbose: true })
    console.info(fs.writeFileSync("js-analysis.log", analysis))
  }

  fs.copyFile(srcDir + 'index.html', outDir + '/index.html', (err) => {
    if (err) throw err
  })
  const staticFiles = fs.readdirSync('./static')
  await Promise.all(
    staticFiles.map((f) => {
      return new Promise((res) => {
        const src = path.join(staticDir, f)
        const target = path.join(outDir, f)
        return fs.copyFile(src, target, res)
      })
    })
  )
}

run()

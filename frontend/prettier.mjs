import { format } from 'prettier'
// This plugin is huge...
import pluginJsDoc from 'prettier-plugin-jsdoc'

process.stdin.on('data', data => {
  const formatted = format(data.toString(), { parser: 'babel-ts', plugins: [pluginJsDoc] })
  console.log(formatted)
  process.exit()
})

import { TemplateConfig } from 'src/common-ts/TemplateConfig'

/**
 * Normally, we would name this enum `TemplateID` but since this enum is exposed to users
 * it makes more sense to name it `Env` because it's less confusing for users.
 */
export enum Env {
  NodeJS = 'nodejs-v16',
}

export const templates: { [key in Env]: TemplateConfig & { command: string } } = {
  'nodejs-v16': {
    id: 'nodejs-v16',
    image: 'us-central1-docker.pkg.dev/devbookhq/devbook-runner-templates/nodejs-v16:latest',
    root_dir: '/home/runner',
    code_cells_dir: '/home/runner/src',
    command: 'node -e ',
  },
}

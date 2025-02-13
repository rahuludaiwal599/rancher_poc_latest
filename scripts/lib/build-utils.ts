/**
 * This module is a helper for the build & dev scripts.
 */

import childProcess from 'child_process';
import fs from 'fs';
import path from 'path';
import util from 'util';

import spawn from 'cross-spawn';
import _ from 'lodash';
import webpack from 'webpack';

import babelConfig from 'babel.config';

/**
 * A promise that is resolved when the child exits.
 */
type SpawnResult = Promise<void> & {
  child: childProcess.ChildProcess;
};

export default {
  /**
   * Determine if we are building for a development build.
   */
  isDevelopment: true,

  get serial() {
    return process.argv.includes('--serial');
  },

  sleep: util.promisify(setTimeout),

  /**
   * Get the root directory of the repository.
   */
  get rootDir() {
    return path.resolve(__dirname, '..', '..');
  },

  get rendererSrcDir() {
    return path.resolve(this.rootDir, `${ process.env.RD_ENV_PLUGINS_DEV ? '' : 'pkg/rancher-desktop' }`);
  },

  /**
   * Get the directory where all of the build artifacts should reside.
   */
  get distDir() {
    return path.resolve(this.rootDir, 'dist');
  },

  /**
   * Get the directory holding the generated files.
   */
  get appDir() {
    return path.resolve(this.distDir, 'app');
  },

  /** The package.json metadata. */
  get packageMeta() {
    const raw = fs.readFileSync(path.join(this.rootDir, 'package.json'), 'utf-8');

    return JSON.parse(raw);
  },

  /**
  * Spawn a new process, returning the child process.
  * @param command The executable to spawn.
  * @param args Arguments to the executable. The last argument may be
  *                        an Object holding options for child_process.spawn().
  */
  spawn(command: string, ...args: any[]): SpawnResult {
    const options: childProcess.SpawnOptions = {
      cwd:   this.rootDir,
      stdio: 'inherit',
    };

    if (args.concat().pop() instanceof Object) {
      Object.assign(options, args.pop());
    }

    const child = spawn(command, args, options);

    const promise: Promise<void> = new Promise((resolve, reject) => {
      child.on('exit', (code, signal) => {
        if (signal && signal !== 'SIGTERM') {
          reject(new Error(`Process exited with signal ${ signal }`));
        } else if (code !== 0 && code !== null) {
          reject(new Error(`Process exited with code ${ code }`));
        }
      });
      child.on('error', (error) => {
        reject(error);
      });
      child.on('close', resolve);
    });

    return Object.assign(promise, { child });
  },

  /**
   * Execute the passed-in array of tasks and wait for them to finish.  By
   * default, all tasks are executed in parallel.  The user may pass `--serial`
   * on the command line to causes the tasks to be executed serially instead.
   * @param tasks Tasks to execute.
   */
  async wait(...tasks: (() => Promise<void>)[]) {
    if (this.serial) {
      for (const task of tasks) {
        await task();
      }
    } else {
      await Promise.all(tasks.map(t => t()));
    }
  },

  /**
   * Get the webpack configuration for the main process.
   */
  get webpackConfig(): webpack.Configuration {
    const mode = this.isDevelopment ? 'development' : 'production';

    const config: webpack.Configuration = {
      mode,
      target: 'electron-main',
      node:   {
        __dirname:  false,
        __filename: false,
      },
      entry:     { background: path.resolve(this.rootDir, 'background') },
      externals: [...Object.keys(this.packageMeta.dependencies)],
      devtool:   this.isDevelopment ? 'source-map' : false,
      resolve:   {
        alias:      { '@pkg': path.resolve(this.rootDir, 'pkg', 'rancher-desktop') },
        extensions: ['.ts', '.js', '.json', '.node'],
        modules:    ['node_modules'],
      },
      output: {
        libraryTarget: 'commonjs2',
        filename:      '[name].js',
        path:          this.appDir,
      },
      module: {
        rules: [
          {
            test: /\.ts$/,
            use:  {
              loader:  'ts-loader',
              options: { transpileOnly: this.isDevelopment },
            },
          },
          {
            test: /\.js$/,
            use:  {
              loader:  'babel-loader',
              options: {
                ...babelConfig,
                cacheDirectory: true,
              },
            },
            exclude: [/node_modules/, this.distDir],
          },
          {
            test:    /\.ya?ml$/,
            exclude: [/(?:^|[/\\])assets[/\\]scripts[/\\]/, this.distDir],
            use:     { loader: 'js-yaml-loader' },
          },
          {
            test: /\.node$/,
            use:  { loader: 'node-loader' },
          },
          {
            test: /(?:^|[/\\])assets[/\\]scripts[/\\]/,
            use:  { loader: 'raw-loader' },
          },
        ],
      },
      plugins: [
        new webpack.EnvironmentPlugin({ NODE_ENV: mode }),
      ],
    };

    return config;
  },

  /**
   * WebPack configuration for the preload script
   */
  get webpackPreloadConfig(): webpack.Configuration {
    const overrides: webpack.Configuration = {
      target: 'electron-preload',
      output: {
        filename: '[name].js',
        path:     path.join(this.rootDir, 'resources'),
      },
    };

    const result = Object.assign({}, this.webpackConfig, overrides);
    const rules = result.module?.rules ?? [];

    const uses = rules.filter(
      (rule): rule is webpack.RuleSetRule => typeof rule !== 'boolean' && typeof rule !== 'string',
    );

    const tsLoader = uses.find(u => u.loader === 'ts-loader');

    if (tsLoader) {
      tsLoader.options = _.merge({}, tsLoader.options, { compilerOptions: { noEmit: false } });
    }

    result.entry = { preload: path.resolve(this.rendererSrcDir, 'preload', 'index.ts') };

    return result;
  },

  /**
   * Build the main process JavaScript code.
   */
  buildJavaScript(config: webpack.Configuration): Promise<void> {
    return new Promise((resolve, reject) => {
      webpack(config).run((err, stats) => {
        if (err) {
          return reject(err);
        }
        if (stats?.hasErrors()) {
          return reject(new Error(stats.toString({ colors: true, errorDetails: true })));
        }
        console.log(stats?.toString({ colors: true }));
        resolve();
      });
    });
  },

  get arch(): NodeJS.Architecture {
    return process.env.M1 ? 'arm64' : process.arch;
  },

  /**
   * Build the preload script.
   */
  async buildPreload(): Promise<void> {
    await this.buildJavaScript(this.webpackPreloadConfig);
  },

  /**
   * Build the main process code.
   */
  buildMain(): Promise<void> {
    return this.wait(() => this.buildJavaScript(this.webpackConfig));
  },

};

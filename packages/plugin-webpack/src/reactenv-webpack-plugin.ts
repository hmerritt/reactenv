import type { Compiler } from 'webpack';

type Definitions = Record<string, any>;

export class ReactEnvReplacementPlugin {
    pluginName: string;
    isBuild: boolean;
    keys: string[];
    defaultValues: Record<string, any>;

    constructor(...keys: (string | string[] | Record<string, any>)[]) {
        this.pluginName = 'ReactEnvReplacementPlugin';
        this.isBuild = false;

        if (keys.length === 1 && Array.isArray(keys[0])) {
            this.keys = keys[0];
            this.defaultValues = {};
        } else if (
            keys.length === 1 &&
            keys[0] &&
            typeof keys[0] === 'object'
        ) {
            this.keys = Object.keys(keys[0]);
            this.defaultValues = keys[0];
        } else {
            this.keys = keys as string[];
            this.defaultValues = {};
        }
    }

    apply(compiler: Compiler) {
        this.isBuild = compiler.options.mode === 'production';

        // Hook into the environment setup phase
        compiler.hooks.thisCompilation.tap(this.pluginName, () => {
            let definitions = {} as Definitions;

            if (this.isBuild) {
                console.debug('[reactenv]', 'mode: production');
                definitions = this.onBuild(definitions);
            } else {
                console.debug('[reactenv]', 'mode: development');
                definitions = this.onDev(definitions, compiler);
            }

            // Add the plugin to the webpack configuration
            const { DefinePlugin } = compiler.webpack;
            new DefinePlugin(definitions).apply(compiler);
        });
    }

    /**
     * Apply production environment variable replacement, used only when building
     * @returns {Record<string, CodeValue>}
     */
    onBuild(definitions: Definitions) {
        const appEnv = Object.keys(process.env).filter(
            (key) =>
                key.startsWith('REACT_') ||
                key.startsWith('VITE_') ||
                key.startsWith('VUE_')
        );

        // Convert our environment variables to the correct format
        appEnv.forEach((key) => {
            definitions[`process.env.${key}`] = JSON.stringify(
                `__reactenv.${key}`
            );
        });

        return definitions;
    }

    /**
     * Apply development environment variables, mimics `webpack.EnvironmentPlugin`
     * @returns {Record<string, CodeValue>}
     */
    onDev(definitions: Definitions, compiler: Compiler) {
        const { webpack } = compiler;

        for (const key of this.keys) {
            const value =
                process.env[key] !== undefined
                    ? process.env[key]
                    : this.defaultValues[key];

            if (value === undefined) {
                compiler.hooks.thisCompilation.tap(
                    this.pluginName,
                    (compilation) => {
                        const error = new webpack.WebpackError(
                            `EnvironmentPlugin - ${key} environment variable is undefined.\n\n` +
                                'You can pass an object with default values to suppress this warning.\n' +
                                'See https://webpack.js.org/plugins/environment-plugin for example.'
                        );

                        error.name = 'EnvVariableNotDefinedError';
                        compilation.errors.push(error);
                    }
                );
            }

            definitions[`process.env.${key}`] =
                value === undefined ? 'undefined' : JSON.stringify(value);
        }

        return definitions;
    }
}

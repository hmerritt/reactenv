const webpack = require("webpack");

class ReactEnvReplacementPlugin {
    constructor() {
        this.pluginName = "ReactEnvReplacementPlugin";
    }

    apply(compiler) {
        const isEnabled = compiler.options.mode === "production";
        console.debug("[reactenv]", "enabled:", isEnabled);
        if (!isEnabled) return;

        compiler.hooks.compilation.tap(this.pluginName, (compilation) => {
            // Use processAssets hook with PROCESS_ASSETS_STAGE_DEV_TOOLING stage
            compilation.hooks.processAssets.tap(
                {
                    name: this.pluginName,
                    stage: webpack.Compilation.PROCESS_ASSETS_STAGE_ADDITIONAL,
                },
                (assets) => {
                    Object.entries(assets).forEach(([filename, asset]) => {
                        if (!this.shouldProcessFile(filename)) return;

                        let content = asset.source();

                        // Replace process.env.X with reactenv.X
                        const envVarRegex = /process\.env\.(\w+)/g;
                        const newContent = content.replace(
                            envVarRegex,
                            (match, varName) => {
                                return `"reactenv.${varName}"`;
                            },
                        );

                        // Only update if content has changed
                        if (newContent !== content) {
                            compilation.updateAsset(
                                filename,
                                new webpack.sources.RawSource(newContent),
                            );
                        }
                    });
                },
            );
        });
    }

    shouldProcessFile(filename) {
        return (
            /\.(js|ts|jsx|tsx)$/.test(filename) && !filename.includes("node_modules")
        );
    }
}

module.exports = ReactEnvReplacementPlugin;

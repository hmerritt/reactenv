const webpack = require("webpack");
const PLUGIN_NAME = "ReactEnvWebpackPlugin";

class ReactEnvWebpackPlugin {
    apply(compiler) {
        // Only run in production mode
        const isEnabled = process.env.NODE_ENV === "production";
        console.debug("[reactenv]", "enabled:", isEnabled);
        // if (!isEnabled) return;

        compiler.hooks.initialize.tap(PLUGIN_NAME, () => {
            // Create replacement definitions
            const definitions = {};
            Object.keys(process.env).forEach((key) => {
                definitions[`process.env.${key}`] = JSON.stringify(`reactenv.${key}`);
            });

            // Apply DefinePlugin with our replacements
            new webpack.DefinePlugin(definitions).apply(compiler);
        });
    }
}

module.exports = ReactEnvWebpackPlugin;

const os = require('os')
const path = require('path')
const fs = require('fs')

function getPlatformPackage() {
	const platform = os.platform()
	const arch = os.arch()

	const platforms = {
		'win32': {
			'x64': 'win32-x64',
			'arm64': 'win32-arm64'
		},
		'darwin': {
			'x64': 'darwin-x64',
			'arm64': 'darwin-arm64'
		},
		'linux': {
			'x64': 'linux-x64',
			'arm64': 'linux-arm64'
		}
	}

	const platformSupport = platforms[platform]
	if (!platformSupport) throw new Error(`Unsupported platform: ${platform}`)

	const binary = platformSupport[arch]
	if (!binary) throw new Error(`Unsupported architecture: ${arch}`)

	return `@hmerritt/reactenv-${binary}`
}

async function install() {
	try {
		const platformPackage = getPlatformPackage()
		const binaryPath = require.resolve(platformPackage)

		const binDir = path.join(__dirname, 'bin')
		if (!fs.existsSync(binDir)) {
			fs.mkdirSync(binDir, { recursive: true })
		}

		// Copy the binary to the bin directory
		const destPath = path.join(binDir, '_reactenv' + (os.platform() === 'win32' ? '.exe' : ''))
		fs.copyFileSync(binaryPath, destPath)

		// Make the binary executable (not needed on Windows)
		if (os.platform() !== 'win32') {
			fs.chmodSync(destPath, 0o755)
		}

		console.log('[reactenv] Binary installed successfully')
	} catch (error) {
		console.error('[reactenv] Failed to install binary:', error)
		process.exit(1)
	}
}

install().catch(err => {
	console.error(err)
	process.exit(1)
})

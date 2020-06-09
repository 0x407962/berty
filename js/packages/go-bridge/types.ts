export enum GoLogLevel {
	debug = 'debug',
	info = 'info',
	error = 'error',
	warn = 'warn',
}

export type GoBridgeOpts = {
	swarmListeners?: string[]
	grpcListeners?: string[]
	persistance?: boolean
	logLevel?: GoLogLevel
	tracing?: boolean
}

export type GoLoggerOpts = {
	level: GoLogLevel
	message: string
}

export interface GoBridgeInterface {
	log(_: GoLoggerOpts): void
	startProtocol(_: GoBridgeOpts): Promise<void>
	stopProtocol(): Promise<void>
	getProtocolAddr(): Promise<string>
	clearStorage(): Promise<void>
}

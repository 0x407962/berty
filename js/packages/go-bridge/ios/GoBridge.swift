//
//  GoBridge.swift
//  GoBridge
//
//  Created by Guilhem Fanton on 06/01/2020.
//  Copyright © 2020 Berty Technologies. All rights reserved.
//

import Foundation
import Bertybridge

@objc(GoBridge)
class GoBridge: NSObject {
    let orbitdir: URL

    var bridgeDemo: BertybridgeDemo?

    override init() {
        // set berty dir for persistance
        let absUserUrl = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask).first!
        self.orbitdir = absUserUrl.appendingPathComponent("orbitdb", isDirectory: true)

        super.init()
    }

    @objc func startDemo(_ opts: NSDictionary, resolve: @escaping RCTPromiseResolveBlock, reject: @escaping RCTPromiseRejectBlock) {
        do {
            if self.bridgeDemo != nil {
                throw NSError(domain: "already started", code: 1)
            }

            // get opts
            let optPersistance = opts["persistance"] as? Bool ?? false
            let optLog = opts["logLevel"] as? String ?? "info"
            let optGrpcListeners = opts["grpcListeners"] as? String ?? "/ip4/127.0.0.1/tcp/0/grpcws"
            let optSwarmListeners = opts["swarmListeners"] as? String ?? "/ip4/0.0.0.0/tcp/0,/ip6/0.0.0.0/tcp/0"

            var err: NSError?
            guard let config = BertybridgeNewDemoConfig() else {
                throw NSError(domain: "unable to create config", code: 1)
            }

            // init logger
            let logger = LoggerDriver("tech.berty", "demo")
            config.logLevel(optLog)
            config.loggerDriver(logger)

            // configure grpc listener
            config.addGRPCListener(optGrpcListeners)

            // configure swarm listeners
            config.swarmListeners(optSwarmListeners)

            // set persistance if needed
            if optPersistance {
                var isDirectory: ObjCBool = true
                let exist = FileManager.default.fileExists(atPath: self.orbitdir.path, isDirectory: &isDirectory)
                if !exist {
                    try FileManager.default.createDirectory(atPath: self.orbitdir.path, withIntermediateDirectories: true, attributes: nil)
                }

                config.orbitDBDirectory(self.orbitdir.path)
            }

            let bridgeDemo = BertybridgeNewDemoBridge(config, &err)
            if err != nil {
                throw err!
            }

            self.bridgeDemo = bridgeDemo

            resolve(true)
        } catch let error as NSError {
            reject("\(String(describing: error.code))", error.userInfo.description, error)
        }
    }

    @objc func getDemoAddr(_ resolve: @escaping RCTPromiseResolveBlock, reject: @escaping RCTPromiseRejectBlock) {
        do {
            guard let bridgeDemo = self.bridgeDemo else {
                throw NSError(domain: "bridgeDemo isn't started", code: 1)
            }

            let addr = bridgeDemo.grpcWebSocketListenerAddr()
            resolve(addr)
        } catch let error as NSError {
            reject("\(String(describing: error.code))", error.userInfo.description, error)
        }
    }
}

//
//  BertyDevice.m
//  ble
//
//  Created by sacha on 03/06/2019.
//  Copyright © 2019 berty. All rights reserved.
//

#import "BertyDevice.h"
#import "ble.h"

CBService *getService(NSArray *services, NSString *uuid) {
    CBService *result = nil;
    
    for (CBService *service in services) {
        if ([[service.UUID UUIDString] isEqual:uuid]) {
            result = service;
        }
    }
    return result;
}

@implementation BertyDevice

- (instancetype)initWithPeripheral:(CBPeripheral *)peripheral
                           central:(BleManager *)manager {
    self = [super init];

    if (self) {
        self.remoteCentral = nil;
        self.peripheral = peripheral;
        peripheral.delegate = self;
        self.manager = manager;
        self.maSend = FALSE;
        self.peerIDSend = FALSE;
        self.maRecv = FALSE;
        self.peerIDRecv = FALSE;

        self.dQueue = dispatch_queue_create([[NSString stringWithFormat:@"%@%@",
                                         @"BertyDevice-",
                                         [peripheral.identifier UUIDString]]
                                               cStringUsingEncoding:NSASCIIStringEncoding],
                                        DISPATCH_QUEUE_SERIAL);
        
        self.writeQueue = dispatch_queue_create([[NSString stringWithFormat:@"%@%@",
                                              @"WriteBertyDevice-",
                                              [peripheral.identifier UUIDString]]
                                             cStringUsingEncoding:NSASCIIStringEncoding],
                                            DISPATCH_QUEUE_SERIAL);
        
        void (^maHandler)(NSData *data) = ^(NSData *data) {
            [self handleMa:data];
        };
        
        void (^peerIDHandler)(NSData *data) = ^(NSData *data) {
            [self handlePeerID:data];
        };

        self.characteristicHandlers = @{
                                        [manager.writerUUID UUIDString]: ^(NSData *data) {
                                           NSLog(@"juste writed");
                                        },
                                        [manager.maUUID UUIDString]: [maHandler copy],
                                        [manager.peerUUID UUIDString]: [peerIDHandler copy],
                                        };

        self.characteristicDatas = @{
                                     [manager.writerUUID UUIDString]: [NSMutableData data],
                                     [manager.maUUID UUIDString]: [NSMutableData data],
                                     [manager.peerUUID UUIDString]: [NSMutableData data],
                                     };
    }

    return self;
}

- (void)handleMa:(NSData *)maData {
    self.remoteMa = [NSString stringWithUTF8String:[maData bytes]];
    self.maRecv = TRUE;
    [self checkAndSendToLibP2P];
    NSLog(@"remote Ma %@", self.remoteMa);
}

- (void)handlePeerID:(NSData *)peerIDData {
    self.remotePeerID = [NSString stringWithUTF8String:[peerIDData bytes]];
    self.peerIDRecv = TRUE;
    [self checkAndSendToLibP2P];
    NSLog(@"remote PeerID %@", self.remotePeerID);
}

- (void)checkAndSendToLibP2P {
    if (self.maSend == TRUE && self.peerIDSend == TRUE &&
        self.maRecv == TRUE && self.peerIDRecv == TRUE) {
        NSLog(@"Send To libp2p");
    }
}

- (void)handleDiscoverServices:(NSArray *)services withError:(NSError *)error {
    
}

- (void)handshake {
    dispatch_async(self.dQueue, ^{
        [self connectWithOptions:nil
            withBlock:^(BertyDevice* device, NSError *error){
            if (error) {
                return;
            }

            [self discoverServices:@[self.manager.serviceUUID] withBlock:^(NSArray *services, NSError *error) {
                if (error) {
                    return;
                }

                CBService *service = getService(services, [self.manager.serviceUUID UUIDString]);
                if (service == nil) {
                    return;
                }
                [self discoverCharacteristics:@[self.manager.maUUID, self.manager.peerUUID, self.manager.writerUUID, self.manager.closerUUID,] forService:service withBlock:^(NSArray *chars, NSError *error) {
                    if (error) {
                        return;
                    }
                    
                    for (CBCharacteristic *chr in chars) {
                        if ([chr.UUID isEqual:self.manager.maUUID]) {
                            self.ma = chr;
                        } else if ([chr.UUID isEqual:self.manager.peerUUID]) {
                            self.peerID = chr;
                        } else if ([chr.UUID isEqual:self.manager.writerUUID]) {
                            self.writer = chr;
                        }
                    }

                    [self writeToCharacteristic:[[self.manager.ma dataUsingEncoding:NSUTF8StringEncoding] mutableCopy] forCharacteristic:self.ma withEOD:TRUE andBlock:^(NSError *error) {
                        if (error) {
                            return;
                        }

                        self.maSend = TRUE;
                        [self writeToCharacteristic:[[self.manager.peerID dataUsingEncoding:NSUTF8StringEncoding] mutableCopy] forCharacteristic:self.peerID withEOD:TRUE andBlock:^(NSError *error) {
                            if (error) {
                                return;
                            }

                            self.peerIDSend = TRUE;
                            [self checkAndSendToLibP2P];
                        }];
                    }];
                }];
            }];
        }];
    });
}

- (void)peripheral:(CBPeripheral *)peripheral didModifyServices:(NSArray<CBService *> *)invalidatedServices {
    CBService *service = getService(invalidatedServices, [self.manager.serviceUUID UUIDString]);
    if (service == nil) {
        return;
    }
    NSLog(@"invalidated");
    self.maSend = FALSE;
    self.peerIDSend = FALSE;
    self.maRecv = FALSE;
    self.peerIDRecv = FALSE;
    
    [self.manager.cManager cancelPeripheralConnection:peripheral];
    // TODO: advertise libp2p that it fail
}

- (void)handleConnect:(NSError *)error {
    _BERTY_ON_D_THREAD(^{
        self.connectCallback(self, error);
        self.connectCallback = nil;
    });
}

- (void)connectWithOptions:(NSDictionary *)options withBlock:(void (^)(BertyDevice *, NSError *))connectCallback {
    _BERTY_ON_D_THREAD(^{
        self.connectCallback = connectCallback;
        [self.manager.cManager connectPeripheral:self.peripheral options:nil];
    });
}

#pragma mark - write functions

- (NSData *)getDataToSend {
    NSData *result = nil;

    if (self.remainingData == nil || self.remainingData.length <= 0) {
        return result;
    }

    NSUInteger chunckSize = self.remainingData.length > [self.peripheral maximumWriteValueLengthForType:CBCharacteristicWriteWithResponse] ? [self.peripheral maximumWriteValueLengthForType:CBCharacteristicWriteWithResponse] : self.remainingData.length;

    result = [NSData dataWithBytes:[self.remainingData bytes] length:chunckSize];
    
    if (self.remainingData.length <= chunckSize) {
        self.remainingData = nil;
    } else {
        [self.remainingData setData:[[NSData alloc]
                                 initWithBytes:[self.remainingData mutableBytes] + chunckSize
                                 length:[self.remainingData length] - chunckSize]];
    }

    return result;
}

- (void)writeToCharacteristic:(NSMutableData *)data forCharacteristic:(CBCharacteristic *)characteristic withEOD:(BOOL)eod andBlock:(void (^)(NSError *))writeCallback {
    dispatch_async(self.writeQueue, ^{
        NSData *toSend = nil;
        __block NSError *blockError = nil;
        
        self.remainingData = data;
        dispatch_semaphore_t sema = dispatch_semaphore_create(0);

        while (self.remainingData.length > 0) {
            toSend = [self getDataToSend];
            
            [self writeValue:toSend forCharacteristic:characteristic withBlock:^(NSError *error){
                blockError = error;
                dispatch_semaphore_signal(sema);
            }];
            dispatch_semaphore_wait(sema, DISPATCH_TIME_FOREVER);
            
            if (blockError != nil) {
                writeCallback(blockError);
            }
        }
        NSLog(@"writed on %@", [characteristic.UUID UUIDString]);
        if (eod) {
            [self writeValue:[@"EOD" dataUsingEncoding:NSUTF8StringEncoding] forCharacteristic:characteristic withBlock:^(NSError *error){
                blockError = error;
                dispatch_semaphore_signal(sema);
            }];
            dispatch_semaphore_wait(sema, DISPATCH_TIME_FOREVER);
            NSLog(@"writed EOD");
        }
        dispatch_release(sema);
        writeCallback(nil);
    });
}

- (void)writeValue:(NSData *)value forCharacteristic:(nonnull CBCharacteristic *)characteristic withBlock:(void (^)(NSError * __nullable))writeCallback {
    _BERTY_ON_D_THREAD(^{
        self.writeCallback = writeCallback;
        [self.peripheral writeValue:value forCharacteristic:characteristic type:CBCharacteristicWriteWithResponse];
    });
}

- (void)peripheral:(CBPeripheral *)peripheral didWriteValueForCharacteristic:(CBCharacteristic *)characteristic error:(NSError *)error {
    _BERTY_ON_D_THREAD(^{
        NSLog(@"writeCallback %@", error);
        self.writeCallback(error);
        self.writeCallback = nil;
    });
}

#pragma mark - Characteristic Discovery

- (void)discoverCharacteristics:(nullable NSArray *)characteristics forService:(CBService *)service withBlock:(void (^)(NSArray *, NSError  *))characteristicCallback {
    _BERTY_ON_D_THREAD(^{
        self.characteristicCallback = characteristicCallback;
        [self.peripheral discoverCharacteristics:characteristics forService:service];
    });
}

- (void)peripheral:(CBPeripheral *)peripheral didDiscoverCharacteristicsForService:(CBService *)service error:(NSError *)error {
    _BERTY_ON_D_THREAD(^{
        self.characteristicCallback(service.characteristics, error);
        self.characteristicCallback = nil;
    });
}

#pragma mark - Services Discovery

- (void)discoverServices:(NSArray *)serviceUUIDs withBlock:(void (^)(NSArray *, NSError *))serviceCallback {
    _BERTY_ON_D_THREAD(^{
        self.serviceCallback = serviceCallback;
        [self.peripheral discoverServices:@[self.manager.serviceUUID]];
    });
}

- (void)peripheral:(CBPeripheral *)peripheral didDiscoverServices:(NSError *)error {
    _BERTY_ON_D_THREAD(^{
        self.serviceCallback(peripheral.services, error);
        self.serviceCallback = nil;
    });
}


@end

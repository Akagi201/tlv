----------------------------------------
-- @file tlv.lua
--
-- @author Akagi201 <akagi201@gmail.com>
--
-- @desc
-- Use this script with wssdl <https://github.com/diacritic/wssdl>
--
----------------------------------------
local wssdl = require 'wssdl'

tlv = wssdl.packet {
    type:u8();
    length:i32();
    data:bytes();
}

wssdl.dissect {
    udp.port:set {
        [8327] = tlv:proto('tlv', "the tlv protocol")
    }
}

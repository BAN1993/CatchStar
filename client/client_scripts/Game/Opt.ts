import { biostream } from "../Protocol/biostream";

const { ccclass, property } = cc._decorator;

export enum OPTTYPE {
  GAME_START = 0,
  PLAYER_LEAVE = 1,
  PLAYER_RECONNECT = 2,
  MOVE = 3,

  CLIENT_BEGIN = 3000,
  //CREATE_MONSTER = 3001,
  //CREATE_TURRET = 3002,
  //TURRET_SHOOT = 3003,
  CLIENT_MOVE = 3004,
  CLIENT_JUMP = 3005,
  NEW_STAR = 3006,
  CATCH_STAR = 3007,
  //CLIENT_POSITION = 3008,
  DEL_STAR = 3009,
}

export class FrameData {
  public frameindex: number = 0
  public optcnt: number = 0
  public optdatas: OptData[] = []

  public read(bio: biostream) {
    this.frameindex = bio.readUint32()
    this.optcnt = bio.readUint32()
    for (let i = 0; i < this.optcnt; i++) {
      let data = new OptData()
      data.read(bio)
      this.optdatas.push(data)
    }
  }
}

export class OptData {
  public numid: number = 0
  public opttype: number = 0
  public datalen: number = 0
  public data: ArrayBuffer = null

  public read(bio: biostream) {
    this.numid = bio.readUint32()
    this.opttype = bio.readUint32()
    this.datalen = bio.readUint32()
    this.data = bio.readArrayBuffer(this.datalen)
  }

  public write(bio: biostream) {
    bio.writeUint32(this.numid)
    bio.writeUint32(this.opttype)
    bio.writeUint32(this.datalen)
    bio.writeArrayBuffer(this.data, this.datalen)
  }
}

export class OptEmpty {

}

export class OptGameStart {
  public playerCnt: number = 0
  public players: OptGameStart_PlayerData[] = []

  public read(bio: biostream) {
    this.playerCnt = bio.readUint32()
    for (let i = 0; i < this.playerCnt; i++) {
      let data = new OptGameStart_PlayerData()
      data.read(bio)
      this.players.push(data)
    }
  }
}

export class OptGameStart_PlayerData {
  public who: number = 0
  public numid: number = 0
  public nickname: string = ""
  public gold: number = 0
  public px: number = 0
  public py: number = 0

  public read(bio: biostream) {
    this.who = bio.readUint16()
    this.numid = bio.readUint32()
    this.nickname = bio.readString()
    this.gold = bio.readUint32()
    this.px = bio.readInt32()
    this.py = bio.readInt32()
  }
}

export class OptMove {
  public aim: number = 0
  public px: number = 0
  public py: number = 0

  public read(bio: biostream) {
    this.aim = bio.readInt32()
    this.px = bio.readInt32()
    this.py = bio.readInt32()
  }

  public write(bio: biostream) {
    bio.writeInt32(this.aim)
    bio.writeInt32(this.px)
    bio.writeInt32(this.py)
  }
}

export class OptJump {
  public px: number = 0
  public py: number = 0

  public read(bio: biostream) {
    this.px = bio.readInt32()
    this.py = bio.readInt32()
  }

  public write(bio: biostream) {
    bio.writeInt32(this.px)
    bio.writeInt32(this.py)
  }
}

export class OptNewStar {
  public index: number = 0
  public px: number = 0
  public py: number = 0
  public duration: number = 0

  public read(bio: biostream) {
    this.index = bio.readUint32()
    this.px = bio.readInt32()
    this.py = bio.readInt32()
    this.duration = bio.readUint32()
  }
}

export class OptCatchStar {
  public index: number = 0

  public read(bio: biostream) {
    this.index = bio.readUint32()
  }

  public write(bio: biostream) {
    bio.writeUint32(this.index)
  }
}

// export class OptPosition {
//   public aim: number = 0
//   public px: number = 0
//   public py: number = 0

//   public read(bio: biostream) {
//     this.aim = bio.readInt32()
//     this.px = bio.readInt32()
//     this.py = bio.readInt32()
//   }

//   public write(bio: biostream) {
//     bio.writeInt32(this.aim)
//     bio.writeInt32(this.px)
//     bio.writeInt32(this.py)
//   }
// }

export class OptDelStar {
  public index: number = 0

  public read(bio: biostream) {
    this.index = bio.readUint32()
  }
}

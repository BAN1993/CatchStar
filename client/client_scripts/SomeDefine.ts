
/**
 * 一些key值
 */
export class KeyDefine {
  public static Key_Node: string = "n"
  public static Key_CTL: string = "ctl"
  public static Key_Me: string = "kme"
  public static Key_Enemey: string = "kenemey"
  public static Key_Star: string = "kstar"
}

/**
 * 游戏区一些控制值
 */
export class GameConfig {
  public static Add_Score: number = 1 // TODO 得多少分应该服务端下发，这里先偷懒
  public static NetMoveFixDis: number = 10 // 同步网络移动触发修复的偏移量，越小跳动越频繁，但更准确
  public static NetFixDuration: number = 0.1 // 修复网络同步动作的时间
}

/**
 * 移动类型值
 */
export class MoveType {
  public static Stop: number = 0
  public static Left: number = -1
  public static Right: number = 1
}

/**
 * 动作Tag值
 */
export class ActionTag {
  public static FixJump: number = 10000 // 修正跳跃用的动作tag TODO 现在设计貌似不合理，要是有多个同名动作怎么办
  public static FixPosition: number = 10001 // 修正位置动作tag
}
module.exports = {
  "input": {
    "dateTime": "1990-01-01 01:00",
    "gender": "male"
  },
  "pillars": [
    {
      "label": "年柱",
      "gan": "己",
      "zhi": "巳",
      "ganWuxing": "土",
      "zhiWuxing": "火",
      "ganTone": "earth",
      "zhiTone": "fire",
      "ganShiShen": "伤官",
      "hideGan": [
        "丙",
        "庚",
        "戊"
      ],
      "zhiShiShen": [
        "比肩",
        "偏财",
        "食神"
      ],
      "xingYun": "临官",
      "diShi": "临官",
      "xunKong": "戌亥",
      "naYin": "大林木",
      "shenSha": [
        "禄神",
        "天厨贵人",
        "亡神",
        "词馆"
      ]
    },
    {
      "label": "月柱",
      "gan": "丙",
      "zhi": "子",
      "ganWuxing": "火",
      "zhiWuxing": "水",
      "ganTone": "fire",
      "zhiTone": "water",
      "ganShiShen": "比肩",
      "hideGan": [
        "癸"
      ],
      "zhiShiShen": [
        "正官"
      ],
      "xingYun": "胎",
      "diShi": "胎",
      "xunKong": "申酉",
      "naYin": "涧下水",
      "shenSha": [
        "天乙贵人",
        "福星贵人"
      ]
    },
    {
      "label": "日柱",
      "gan": "丙",
      "zhi": "寅",
      "ganWuxing": "火",
      "zhiWuxing": "木",
      "ganTone": "fire",
      "zhiTone": "wood",
      "ganShiShen": "日主",
      "hideGan": [
        "甲",
        "丙",
        "戊"
      ],
      "zhiShiShen": [
        "偏印",
        "比肩",
        "食神"
      ],
      "xingYun": "长生",
      "diShi": "长生",
      "xunKong": "戌亥",
      "naYin": "炉中火",
      "shenSha": [
        "红艳",
        "劫煞",
        "披麻"
      ]
    },
    {
      "label": "时柱",
      "gan": "己",
      "zhi": "丑",
      "ganWuxing": "土",
      "zhiWuxing": "土",
      "ganTone": "earth",
      "zhiTone": "earth",
      "ganShiShen": "伤官",
      "hideGan": [
        "己",
        "癸",
        "辛"
      ],
      "zhiShiShen": [
        "伤官",
        "正官",
        "正财"
      ],
      "xingYun": "养",
      "diShi": "养",
      "xunKong": "午未",
      "naYin": "霹雳火",
      "shenSha": [
        "太极贵人",
        "国印贵人",
        "华盖",
        "词馆"
      ]
    }
  ],
  "wuxingRows": [
    {
      "label": "木",
      "tone": "wood",
      "count": 1,
      "percent": 12.5
    },
    {
      "label": "火",
      "tone": "fire",
      "count": 3,
      "percent": 37.5
    },
    {
      "label": "土",
      "tone": "earth",
      "count": 3,
      "percent": 37.5
    },
    {
      "label": "金",
      "tone": "metal",
      "count": 0,
      "percent": 0
    },
    {
      "label": "水",
      "tone": "water",
      "count": 1,
      "percent": 12.5
    }
  ],
  "wuxingTotal": 8,
  "mingGe": "伤官格",
  "mingGeDesc": "伤官格是八字中最具个性的格局。月令透出伤官，代表命主才华横溢、思维活跃，具有超强的创造力与表达欲，但也往往不服管束、傲气十足。伤官格之人适合艺术、科技、法律、演艺等需要个人风格的领域。伤官见官（正官）在命中会产生冲突，轻则影响仕途，重则招惹官非，需格外留意。伤官配印，则才华得以发挥，成就斐然。",
  "stemGuidance": {
    "primary_favorable": [],
    "secondary_favorable": [
      {
        "stem": "甲",
        "element": "木",
        "ten_god": "偏印",
        "source_layers": [
          "扶抑"
        ],
        "detail": "木属于扶抑喜用，作为天干级可用方向。"
      },
      {
        "stem": "乙",
        "element": "木",
        "ten_god": "正印",
        "source_layers": [
          "扶抑"
        ],
        "detail": "木属于扶抑喜用，作为天干级可用方向。"
      },
      {
        "stem": "丙",
        "element": "火",
        "ten_god": "比肩",
        "source_layers": [
          "扶抑"
        ],
        "detail": "火属于扶抑喜用，作为天干级可用方向。"
      },
      {
        "stem": "丁",
        "element": "火",
        "ten_god": "劫财",
        "source_layers": [
          "扶抑"
        ],
        "detail": "火属于扶抑喜用，作为天干级可用方向。"
      }
    ],
    "conditioning_only": [
      {
        "stem": "壬",
        "element": "水",
        "ten_god": "七杀",
        "source_layers": [
          "日干调候"
        ],
        "detail": "日干调候取壬，但水属于扶抑忌神；用于说明原局调候结构，不作为后天通用喜神。"
      },
      {
        "stem": "戊",
        "element": "土",
        "ten_god": "食神",
        "source_layers": [
          "日干调候"
        ],
        "detail": "日干调候取戊，但土属于扶抑忌神；用于说明原局调候结构，不作为后天通用喜神。"
      }
    ],
    "adverse": [
      {
        "stem": "癸",
        "element": "水",
        "ten_god": "正官",
        "source_layers": [
          "扶抑"
        ],
        "detail": "水属于扶抑忌神，后天遇此天干宜结合全局谨慎判断。"
      },
      {
        "stem": "庚",
        "element": "金",
        "ten_god": "偏财",
        "source_layers": [
          "扶抑"
        ],
        "detail": "金属于扶抑忌神，后天遇此天干宜结合全局谨慎判断。"
      },
      {
        "stem": "辛",
        "element": "金",
        "ten_god": "正财",
        "source_layers": [
          "扶抑"
        ],
        "detail": "金属于扶抑忌神，后天遇此天干宜结合全局谨慎判断。"
      },
      {
        "stem": "己",
        "element": "土",
        "ten_god": "伤官",
        "source_layers": [
          "扶抑"
        ],
        "detail": "土属于扶抑忌神，后天遇此天干宜结合全局谨慎判断。"
      }
    ]
  },
  "personalDay": {
    "dayBranch": "寅",
    "fuyiYongshen": "木火",
    "fuyiJishen": "水金土"
  }
}
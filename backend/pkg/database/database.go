package database

import (
	"database/sql"
	"log"
	"yuanju/configs"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	var err error
	DB, err = sql.Open("postgres", configs.AppConfig.DatabaseURL)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("数据库 Ping 失败: %v", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)

	log.Println("✅ 数据库连接成功")
}

func Migrate() {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		nickname VARCHAR(100),
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS bazi_charts (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID REFERENCES users(id) ON DELETE SET NULL,
		birth_year INTEGER NOT NULL,
		birth_month INTEGER NOT NULL,
		birth_day INTEGER NOT NULL,
		birth_hour INTEGER NOT NULL,
		gender VARCHAR(10) NOT NULL DEFAULT 'male',
		year_gan VARCHAR(10),
		year_zhi VARCHAR(10),
		month_gan VARCHAR(10),
		month_zhi VARCHAR(10),
		day_gan VARCHAR(10),
		day_zhi VARCHAR(10),
		hour_gan VARCHAR(10),
		hour_zhi VARCHAR(10),
		wuxing JSONB,
		dayun JSONB,
		yongshen VARCHAR(20),
		jishen VARCHAR(20),
		chart_hash VARCHAR(64) UNIQUE,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS ai_reports (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		chart_id UUID REFERENCES bazi_charts(id) ON DELETE CASCADE,
		content TEXT NOT NULL,
		model VARCHAR(50),
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_bazi_charts_user_id ON bazi_charts(user_id);
	CREATE INDEX IF NOT EXISTS idx_bazi_charts_hash ON bazi_charts(chart_hash);
	CREATE INDEX IF NOT EXISTS idx_ai_reports_chart_id ON ai_reports(chart_id);

	CREATE TABLE IF NOT EXISTS admins (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		name VARCHAR(100),
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS llm_providers (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(100) NOT NULL,
		type VARCHAR(50) NOT NULL,
		base_url VARCHAR(500) NOT NULL,
		model VARCHAR(100) NOT NULL,
		api_key_encrypted TEXT NOT NULL,
		active BOOLEAN NOT NULL DEFAULT false,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS ai_requests_log (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		chart_id UUID REFERENCES bazi_charts(id) ON DELETE SET NULL,
		provider_id UUID REFERENCES llm_providers(id) ON DELETE SET NULL,
		model VARCHAR(100),
		duration_ms INTEGER,
		status VARCHAR(20) NOT NULL DEFAULT 'success',
		error_msg TEXT,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_ai_requests_log_created ON ai_requests_log(created_at);
	CREATE INDEX IF NOT EXISTS idx_ai_requests_log_provider ON ai_requests_log(provider_id);

	CREATE TABLE IF NOT EXISTS celebrity_records (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(255) NOT NULL,
		gender VARCHAR(10),
		traits TEXT,
		career VARCHAR(255),
		active BOOLEAN NOT NULL DEFAULT true,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_celebrity_records_active ON celebrity_records(active);

	CREATE TABLE IF NOT EXISTS ai_prompts (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		module VARCHAR(50) UNIQUE NOT NULL,
		content TEXT NOT NULL,
		description VARCHAR(255),
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS ai_liunian_reports (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		chart_id UUID REFERENCES bazi_charts(id) ON DELETE CASCADE,
		target_year INTEGER NOT NULL,
		dayun_ganzhi VARCHAR(10),
		content_structured JSONB,
		model VARCHAR(50),
		created_at TIMESTAMPTZ DEFAULT NOW(),
		UNIQUE(chart_id, target_year)
	);
	CREATE INDEX IF NOT EXISTS idx_ai_liunian_reports_chart_id ON ai_liunian_reports(chart_id);

	CREATE TABLE IF NOT EXISTS algo_config (
		key VARCHAR(100) PRIMARY KEY,
		value TEXT NOT NULL,
		description TEXT,
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS algo_tiaohou (
		day_gan VARCHAR(10) NOT NULL,
		month_zhi VARCHAR(10) NOT NULL,
		xi_elements TEXT NOT NULL,
		text TEXT NOT NULL DEFAULT '',
		updated_at TIMESTAMPTZ DEFAULT NOW(),
		PRIMARY KEY (day_gan, month_zhi)
	);
	`

	if _, err := DB.Exec(schema); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 插入默认的流年 Prompt 预设
	defaultLiunianPrompt := `你是一个权威的八字命理大师。该用户的原局分析：
{{.NatalAnalysisLogic}}

目前正行【{{.CurrentDayunGanZhi}}】大运（干十神={{.CurrentDayunGanShiShen}} 支十神={{.CurrentDayunZhiShiShen}}）。
请为他详细批断【{{.TargetYear}} {{.TargetYearGanZhi}}流年】运程（流年干十神={{.TargetYearGanShiShen}} 支十神={{.TargetYearZhiShiShen}}）。

要求直接输出以下JSON结构，不要包含多余Markdown标记：
{
  "career": "事业财运分析（不少于150字）",
  "romance": "感情桃花分析（不少于150字）",
  "health": "健康风险与预警（不少于150字）",
  "advice": "年度锦囊（一句话点睛）"
}`

	insertPromptSQL := `
	INSERT INTO ai_prompts (module, content, description)
	VALUES ('liunian', $1, '流年运势分析（二分流年节点动态调用）')
	ON CONFLICT (module) DO NOTHING;`
	if _, err := DB.Exec(insertPromptSQL, defaultLiunianPrompt); err != nil {
		log.Printf("初始化默认 Prompt 失败: %v", err)
	}

	// 插入默认的过往事件推算 Prompt
	defaultPastEventsPrompt := `你是一位权威的八字命理批断师，擅长以流年干支与原局互动推断过往人生事件。

命主信息：
- 性别：{{.Gender}}
- 日干：{{.DayGan}}
- 原局四柱与十神概要：{{.NatalSummary}}
- 用神/忌神：{{.YongshenInfo}}
- 原局格局：{{.GejuSummary}}
- 加权身强弱：{{.StrengthDetail}}

命主过往大运列表：
{{.DayunList}}

大运合化标签（仅列已构成天干五合的大运）：
{{.DayunHuahe}}

以下是命主过往各流年的命理信号分析（JSON格式，每条信号含 type / evidence / polarity / source）：
{{.YearsData}}

撰写要求：
1. **基底色优先**：每个流年若包含 type=用神基底 信号，应以其 polarity（吉/凶/中性）为整年定调；个别强 polarity 信号（如神煞、伏吟反吟）可在 narrative 中作为转折点叙述，但全年总体倾向必须遵循基底色
2. **神煞强烈**：若信号 source=神煞，按神煞名直接在 narrative 中点出（如"天乙贵人临运""羊刃刃伤"），不要泛泛而谈
3. **合化整段定向**：若大运被列入合化标签，撰写该段 dayun_summaries 时必须点出合化方向（如"丁壬合化木——日主性向偏木，整段大运转向文教/创意"）
4. **加权身强弱**：在 dayun_summaries 与极强/极弱年份的 narrative 中，应结合命主身强弱明细（如"身弱财多，本年财来财去"或"身极强，宜泄不宜帮"）
5. **冲突时**：若单条信号 polarity 与基底色冲突（如财星透干但财为忌神，basis=凶），narrative 必须用转折语点出（如"虽现财星，然为忌神，反主破耗"），避免直白写"财运提升"
6. 每年 narrative 2-3 句中文，无信号年份写"该年运势较为平稳"
7. dayun_summaries 包含：themes（2-4 个主题词）+ summary（80-120 字）
8. 语气权威客观，使用命理术语，不写鸡汤励志语句
9. 直接输出以下 JSON 格式，不要包含多余 Markdown 标记：
{
  "years": [
    {
      "year": 年份数字,
      "age": 虚岁数字,
      "gan_zhi": "干支",
      "dayun_gan_zhi": "大运干支",
      "signals": ["信号类型列表"],
      "narrative": "2-3句批断文字"
    }
  ],
  "dayun_summaries": [
    {
      "gan_zhi": "大运干支",
      "themes": ["主题词1", "主题词2"],
      "summary": "80-120字大运整体总结"
    }
  ]
}`

	if _, err := DB.Exec(
		`INSERT INTO ai_prompts (module, content, description) VALUES ($1, $2, $3) ON CONFLICT (module) DO NOTHING`,
		"past_events", defaultPastEventsPrompt, "过往年份重大事件推算（算法信号+AI组织语言）",
	); err != nil {
		log.Printf("初始化 past_events Prompt 失败: %v", err)
	}
	// 升级旧版（含 dayun_summaries 但不含 YongshenInfo 的） → 新版模板。
	// 已包含 YongshenInfo 的 prompt 不被覆盖（保护 admin 自定义）。
	if _, err := DB.Exec(
		`UPDATE ai_prompts SET content = $1
         WHERE module = 'past_events'
           AND content NOT LIKE '%YongshenInfo%'`,
		defaultPastEventsPrompt,
	); err != nil {
		log.Printf("升级 past_events Prompt 失败: %v", err)
	}

	// 初始化命理知识库模块（4 个 kb_* 模块，ON CONFLICT DO NOTHING 保留用户自定义修改）
	kbSeeds := []struct {
		module      string
		description string
		content     string
	}{
		{
			module:      "kb_shishen",
			description: "十神断事口诀（子平基础定义）",
			content: `十神基础定义与断事要点（请严格遵循此定义，不得自行创造新的十神含义）：

比肩：与日主同五行同阴阳。主自我、兄弟、竞争、固执、合伙。旺则争财克妻；弱则互助为用。
劫财：与日主同五行异阴阳。主劫夺、冒险、义气、破财、伤官变现。
食神：日主所生，同阴阳。主福寿、才艺、子息、衣食无忧；女命旺则克夫制杀有力。
伤官：日主所生，异阴阳。主才华横溢、叛逆、口舌、伤官见官必有灾；旺则克正官。
偏财：日主所克，同阴阳。主偏财、父星（男）、女朋友（男命）、异性缘、投机生意。
正财：日主所克，异阴阳。主正财、稳定收入、妻星（男）、诚信务实。
七杀：克日主，同阴阳。主权威、压力、竞争者、官司灾厄；有制（食神制）则为权贵武将。
正官：克日主，异阴阳。主名誉、规范、官位、丈夫星（女）；旺而有力主贵，过旺则束缚。
偏印：生日主，同阴阳。主偏门学问、宗教、直觉、孤独；过旺则克食神，不利子息与才艺发挥。
正印：生日主，异阴阳。主慈悲、学业、资格证、母星代表；旺则利于学业和文职，过旺则懒散依赖。`,
		},
		{
			module:      "kb_gejv",
			description: "格局判断规则（《子平真诠》格局派核心）",
			content: `格局判断规则（《子平真诠》核心要义）：

一、定格原则
1. 以月令为首要，看月令所透天干定格（月令支藏干，透天干者为格）。
2. 月令无透干时，取月令本气所对应的六亲为格（如寅月本气甲木，日主非甲则看甲对日主的十神关系）。
3. 正格八格：正官格、财格（正偏财）、印格（正偏印）、食神格、伤官格、七杀格（偏官格）。

二、格局高低判断
1. 成格：格局用神得力（有生有扶），喜神透干或藏支，忌神被制化。格局成则主贵。
2. 破格：忌神透干无制，或用神被冲克，或格局混杂。
3. 救格：格局被破但有救神解危，仍可平稳，只是稍减吉力。

三、用神取法
1. 格局定后，依格取用：官格取财印为喜，杀格取食神制化为喜，食神格取财星泄秀为喜。
2. 四柱结构需看整体，不单看某一柱，要论生克制化的来龙去脉。

四、格局与大运的互动
1. 大运走入喜神运：格局由弱转强，事业、财运、婚姻均得力。
2. 大运走入忌神运：格局受冲，应期灾难、变化、损失。
3. 流年与大运叠加：双重喜神叠加为大吉，双重忌神叠加为大凶应期。`,
		},
		{
			module:      "kb_tiaohou",
			description: "调候用神应用指南（十天干大运行运格）",
			content: `调候用神应用指南（十天干大运行运格，梁湘润实战体系）：

调候核心：先论寒暖燥湿，凡命局调候用神透干得地，则体质健旺、运途顺达；调候用神缺失或被冲克，则体弱多病、运势多阻。本段为总纲，具体月令精算数据已在「调候用神精算注入」段落动态提供，断命时请优先参照精算结果。

【甲木】用神：庚金（劈甲）、丙火（暖局）、癸水（滋润）。
大运警示：乙运合庚失去劈甲之力，财运受阻犹豫纠葛；庚辛申酉运主财源极佳；壬癸运寅月无大碍但忌流年刑冲。

【乙木】用神：丙火（暖照）、癸水（润木）、戊土（燥湿）。
大运警示：庚运克乙逢重大损失；辛运合乙主优柔寡断；丙运春月是非纠纷；戊运秋月缺水大凶。

【丙火】用神：壬水（济火）、庚辛（生壬源）、甲木（引火）。
大运警示：壬运秋月宜艺术宗教；癸运春夏进退不一；戊运厚土埋火晦暗；己运四柱无木则财运一落千丈。

【丁火】用神：甲木（引丁）、庚金（劈甲）、壬水（监溺）。
大运警示：壬运合丁事业动荡，破财血光（尤其乙酉乙丑年）；庚运克甲引力失根；乙运合庚丁火有根得用。

【戊土】用神：甲木（疏辟）、丙火（暖照）、癸水（润泽）。
大运警示：己运午月缺水者有寿终之象，其余月令事业一落千丈；庚运亥月孤立无助；壬运亥月虚名虚利；癸运子丑月退财难成就。

【己土】用神：丙火（暖田）、癸壬水（润泽）、甲木（疏通）。
大运警示：戊运夏秋六月不宜，正印格易破产，偏财格防血光官讼；己运感情退位财运失利；壬运辰月易引发诉讼是非；癸运亥子丑月退财难成。

【庚金】用神：丁火（煅炼）、甲木（引丁）、壬水（洗淘）。
大运警示：己运多月令百事无成，亦主婚姻纠葛；乙运合庚犹豫大失意；壬运宜艺术宗教不宜商贾；癸运主病灾寿元忧；午月己运家业一落千丈。

【辛金】用神：壬水（洗照）、丙火（暖局冬月）、己土（生金夏月）。
大运警示：庚运比劫主重大损失或健康事故；丙运春夏是非诉讼频发；戊运怀才不遇；己运官非失职；辛运操之过急必败；戊运申月自毁前程。

【壬水】用神：戊土（堤防）、庚辛（发源）、丁火（制金）。
大运警示：丁运偏印格大破败，七杀格破产之忧；己运伤官格招祸端；甲运得财时体弱多病；癸运偏财格所谋虚浮；戊运酉戌月名利皆失。

【癸水】用神：辛金（发源）、丙火（暖局）、丁火（制庚辛）。
大运警示：丁运多月令家庭感情不利，子丑月等同桃花引祸；丙运午未月先成后败；戊运进退两难或有病患；壬运申月失业闲职；辛运依附他人方能自立。`,
		},
		{
			module:      "kb_yingqi",
			description: "流年应期推算口诀（冲合刑害定月份）",
			content: `流年应期推算方法（定事件发生月份）：

一、流月干支六亲法
流年批断中，需进一步推算应期月份：
- 吉事应期：喜用神旺之月、三合局成之月、贵人临门之月（天乙贵人所在月份）。
- 凶事应期：忌神透出之月、六冲之月（如日支子、流月午为冲，为凶事应期）、刑害之月。

二、六冲应期
子午冲、丑未冲、寅申冲、卯酉冲、辰戌冲、巳亥冲。
- 日主或用神地支被流月地支冲：此月为动荡应期，感情、工作、财运均有波动。
- 喜神被冲：此月凶事发生。忌神被冲：此月凶事反而被化解（吉）。

三、三合、六合成局应期
- 流月地支与年柱、日柱形成三合局（如亥卯未→木局）：木旺之事在此月应验（木主文书、贵人、财缘等视十神而定）。
- 六合之月（子丑合、寅亥合等）：主和合、婚姻、合作事宜成就。

四、具体断法示例
- 事业升迁：正官、偏财旺之月，三合食伤生财之月。
- 桃花感情：桃花星（子午卯酉）临流月，且与日支合者为应期。
- 健康灾厄：病符星行至、六冲日柱或身宫之月，为就医应期。
- 财运：偏正财透出或三合财局之月，为进财应期。`,
		},
		{
			module:      "kb_tonality",
			description: "语调与立场（中立理性分析风格指引）",
			content: `命理分析语调与立场规范：

一、核心原则：如实批断，不偏不倚
你是一位理性的命理分析师，职责是如实呈现命盘信息，而非安慰或讨好用户。

- 好的就说好，差的就说差，有困难的直接讲困难。
- 禁止在每段结尾添加"安慰式总结"或"正能量兜底"。
- 禁止使用"虽然……但只要努力就能……"的万能句式来淡化负面信息。
- 禁止把明显的凶象弱化为"小挑战"或"成长机会"。

二、语气标准
- 使用专业、克制、直接的语气，如同医生对病人做体检报告。
- 遇到命局优势：客观说明优势所在及适用场景，不夸大。
- 遇到命局劣势：明确指出风险、困境及其成因，不回避。
- 遇到矛盾格局：如实说明矛盾（如"格局与调候冲突"），不强行圆说。

三、禁用词与禁用句式
- 禁用词：温馨、暖心、温暖的力量、星光指引、只要你努力、相信自己、命运掌握在自己手中
- 禁用句式：
  × "虽然XX不太理想，但这恰恰是你成长的契机"
  × "命运给了你挑战，也给了你翅膀"
  × "只要把握好方向，一切都会好起来"
  × 任何以鸡汤格言结尾的段落

四、正确示范
- ✓ "日主失令于月柱七杀，身弱无依，事业初期阻力较大，贵人运薄，需依靠自身技艺立足。"
- ✓ "财星被劫，偏财运不佳，投资理财需极度谨慎，此命不宜冒险求财。"
- ✓ "桃花旺但官杀混杂，感情经历丰富但婚姻稳定性堪忧，晚婚为宜。"
- ✓ "此步大运忌神透出，事业财运均有下滑风险，宜守不宜攻。"`,
		},
	}

	for _, seed := range kbSeeds {
		if _, err := DB.Exec(
			`INSERT INTO ai_prompts (module, content, description) VALUES ($1, $2, $3) ON CONFLICT (module) DO NOTHING`,
			seed.module, seed.content, seed.description,
		); err != nil {
			log.Printf("初始化知识模块 [%s] 失败: %v", seed.module, err)
		}
	}

	// 增量迁移 (tiaohou-dayun-sync)：将 kb_tiaohou 更新为覆盖十天干实战大运征兆版本
	newKbTiaohou := `调候用神应用指南（十天干大运行运格，梁湘润实战体系）：

调候核心：先论寒暖燥湿，凡命局调候用神透干得地，则体质健旺、运途顺达；调候用神缺失或被冲克，则体弱多病、运势多阻。本段为总纲，具体月令精算数据已在「调候用神精算注入」段落动态提供，断命时请优先参照精算结果。

【甲木】用神：庚金（劈甲）、丙火（暖局）、癸水（滋润）。
大运警示：乙运合庚失去劈甲之力，财运受阻犹豫纠葛；庚辛申酉运主财源极佳；壬癸运寅月无大碍但忌流年刑冲。

【乙木】用神：丙火（暖照）、癸水（润木）、戊土（燥湿）。
大运警示：庚运克乙逢重大损失；辛运合乙主优柔寡断；丙运春月是非纠纷；戊运秋月缺水大凶。

【丙火】用神：壬水（济火）、庚辛（生壬源）、甲木（引火）。
大运警示：壬运秋月宜艺术宗教；癸运春夏进退不一；戊运厚土埋火晦暗；己运四柱无木则财运一落千丈。

【丁火】用神：甲木（引丁）、庚金（劈甲）、壬水（监溺）。
大运警示：壬运合丁事业动荡，破财血光（尤其乙酉乙丑年）；庚运克甲引力失根；乙运合庚丁火有根得用。

【戊土】用神：甲木（疏辟）、丙火（暖照）、癸水（润泽）。
大运警示：己运午月缺水者有寿终之象，其余月令事业一落千丈；庚运亥月孤立无助；壬运亥月虚名虚利；癸运子丑月退财难成就。

【己土】用神：丙火（暖田）、癸壬水（润泽）、甲木（疏通）。
大运警示：戊运夏秋六月不宜，正印格易破产，偏财格防血光官讼；己运感情退位财运失利；壬运辰月易引发诉讼是非；癸运亥子丑月退财难成。

【庚金】用神：丁火（煅炼）、甲木（引丁）、壬水（洗淘）。
大运警示：己运多月令百事无成，亦主婚姻纠葛；乙运合庚犹豫大失意；壬运宜艺术宗教不宜商贾；癸运主病灾寿元忧；午月己运家业一落千丈。

【辛金】用神：壬水（洗照）、丙火（暖局冬月）、己土（生金夏月）。
大运警示：庚运比劫主重大损失或健康事故；丙运春夏是非诉讼频发；戊运怀才不遇；己运官非失职；辛运操之过急必败；戊运申月自毁前程。

【壬水】用神：戊土（堤防）、庚辛（发源）、丁火（制金）。
大运警示：丁运偏印格大破败，七杀格破产之忧；己运伤官格招祸端；甲运得财时体弱多病；癸运偏财格所谋虚浮；戊运酉戌月名利皆失。

【癸水】用神：辛金（发源）、丙火（暖局）、丁火（制庚辛）。
大运警示：丁运多月令家庭感情不利，子丑月等同桃花引祸；丙运午未月先成后败；戊运进退两难或有病患；壬运申月失业闲职；辛运依附他人方能自立。`

	if _, err := DB.Exec(
		`UPDATE ai_prompts SET content = $1, description = $2, updated_at = NOW() WHERE module = 'kb_tiaohou'`,
		newKbTiaohou,
		"调候用神应用指南（十天干大运行运格）",
	); err != nil {
		log.Printf("增量迁移 (tiaohou-dayun-sync) 失败: %v", err)
	} else {
		log.Println("✅ kb_tiaohou 已同步更新至大运实战征兆版本")
	}

	// 增量迁移：为 ai_reports 表新增 content_structured 字段（JSONB，历史兼容）
	alterSQL := `ALTER TABLE ai_reports ADD COLUMN IF NOT EXISTS content_structured JSONB;`
	if _, err := DB.Exec(alterSQL); err != nil {
		log.Fatalf("增量迁移失败 (content_structured): %v", err)
	}

	// 增量迁移：chart_hash 从全局唯一改为每用户唯一（此段历史增量旧锁已被 resource-based-bazi 废弃取消，防止启动重试添加时遭现存无约束重复记录报错）
	chartHashMigrations := []string{
		`ALTER TABLE bazi_charts DROP CONSTRAINT IF EXISTS bazi_charts_chart_hash_key;`,
		`DROP INDEX IF EXISTS idx_bazi_charts_hash;`,
	}
	for _, migSQL := range chartHashMigrations {
		if _, err := DB.Exec(migSQL); err != nil {
			log.Fatalf("增量迁移失败 (chart_hash isolation): %v\nSQL: %s", err, migSQL)
		}
	}

	// 增量迁移 (resource-based-bazi)：切底解除基于 chart_hash 的独立限制，让每次查命均落位新快照记录
	resourceUnbindMigrations := []string{
		`ALTER TABLE bazi_charts DROP CONSTRAINT IF EXISTS bazi_charts_chart_hash_user_id_key;`,
		`ALTER TABLE bazi_charts DROP CONSTRAINT IF EXISTS bazi_charts_chart_hash_key;`,
		`DROP INDEX IF EXISTS idx_bazi_charts_hash_user;`,
		`DROP INDEX IF EXISTS idx_bazi_charts_hash;`,
	}
	for _, sql := range resourceUnbindMigrations {
		if _, err := DB.Exec(sql); err != nil {
			log.Fatalf("增量迁移失败 (resource unbind migrations): %v\nSQL: %s", err, sql)
		}
	}

	// 增量迁移 (fix-lunar-calendar-persistence)：持久化历法类型与闰月标识，确保历史重排一致性
	calendarMigrations := []string{
		`ALTER TABLE bazi_charts ADD COLUMN IF NOT EXISTS calendar_type VARCHAR(10) NOT NULL DEFAULT 'solar';`,
		`ALTER TABLE bazi_charts ADD COLUMN IF NOT EXISTS is_leap_month BOOLEAN NOT NULL DEFAULT false;`,
	}
	for _, migSQL := range calendarMigrations {
		if _, err := DB.Exec(migSQL); err != nil {
			log.Fatalf("增量迁移失败 (calendar persistence): %v\nSQL: %s", err, migSQL)
		}
	}

	// 增量迁移 (chart-result-snapshot)：持久化整个 BaziResult 快照，避免下游重复调用 lunar-go 与算法漂移
	resultSnapshotMigration := `ALTER TABLE bazi_charts ADD COLUMN IF NOT EXISTS result_json JSONB;`
	if _, err := DB.Exec(resultSnapshotMigration); err != nil {
		log.Fatalf("增量迁移失败 (result_json snapshot): %v", err)
	}

	// 增量迁移 (shensha-annotations)：神煞注解表
	shenshaAnnotationMigration := `
	CREATE TABLE IF NOT EXISTS shensha_annotations (
		id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name        VARCHAR(20)  NOT NULL UNIQUE,
		polarity    VARCHAR(10)  NOT NULL DEFAULT 'zhong',
		description TEXT         NOT NULL DEFAULT '',
		updated_at  TIMESTAMPTZ  DEFAULT NOW()
	);`
	if _, err := DB.Exec(shenshaAnnotationMigration); err != nil {
		log.Fatalf("增量迁移失败 (shensha_annotations): %v", err)
	}

	// 预置神煞文案 seed（ON CONFLICT DO NOTHING 保留管理员自定义修改）
	shenshaSeedData := []struct {
		name        string
		polarity    string
		description string
	}{
		{"天乙贵人", "ji", "天乙贵人为诸神煞之首，乃十二神煞中最为尊贵的吉神。其起源于古代星象，代表上天赐予的庇佑与转化之力。命盘中若日干、年干所对应的地支出现于四柱，即为天乙贵人临门。命带此星者，往往逢凶化吉，险中得救，在人生关键节点总能得到贵人相扶。无论困境、官讼、病灾，只要天乙贵人入局，多半能峰回路转。此星利于从事政界、商界与服务业，亦主聪明伶俐、才思敏捷，深受上位者赏识。双天乙者贵气倍增，但需注意格局是否承受得住，否则会情绪起伏，难以沉稳。"},
		{"太极贵人", "ji", "太极贵人以年干或日干推命，若所对应地支出现于四柱，即为太极贵人入命。此星主人聪慧机敏，善于洞察玄机，具有超凡的悟性与灵感，对哲学、医学、命理、宗教等玄学领域有天然亲近感。命带太极贵人者，往往悟性极高，处事圆融，凡事能从全局出发，多谋善断。此星亦主长寿与福德，人生中常有意想不到的转机。在学术研究、策划谋略、文艺创作领域尤为有利。贵人之地若逢生旺，则一生贵气浓厚，遇难成祥；若临衰地，则聪明反被聪明误，需引以为戒。"},
		{"文昌贵人", "ji", "文昌贵人以日干所对应的文昌支推命，若该支出现于四柱，即为文昌入局。文昌主聪明才学，是命理中最具文化气质的吉星。命带文昌者，自幼好学，记忆力强，思维敏捷，擅长文字、语言与艺术，在学业考试中往往发挥出色，适合从事教育、文学、传媒、法律和各类文职工作。此星亦主主人气质优雅，谦逊有礼，能赢得周围人的好感与欣赏。文昌入日柱或时柱时，文学才华尤为突出；若见空亡或被刑冲，则聪明难以持续发挥，需通过后天用功加以弥补。"},
		{"禄神", "ji", "禄神又称建禄、临官，是日干在地支达到临官旺地的神煞，象征衣食无忧、自食其力、职业稳定的格局标志。命带禄神者，天生具有独立谋生的能力，工作积极进取，不依赖他人施舍，凭借自身努力获取财富。古籍云：「禄者，食也」，此星主人一生不愁吃穿，温饱有余。若禄神透天干或坐日柱，则自立能力极强，事业心旺盛，往往能在所从事的领域中建立一定地位。禄神旺者，健康情况亦较为良好。然禄多不宜见财，禄旺格局重者，需留意是否有劫财兄弟争夺。"},
		{"天德贵人", "ji", "天德贵人是以月支推算的德星，不同月份对应不同天干或地支，若四柱中出现对应的干支，即为命带天德。此星代表上天赐予的德行庇佑，是解厄化煞的强力吉神。命带天德者，心性善良，行事正直，有侠义精神，往往能在灾难来临前得到神助或贵人预警，逢凶化吉。在法律纠纷、病灾意外、事业失利等关键时刻，天德贵人能发挥化险为夷的作用。此星亦主聪慧仁厚，具有领导气质，受人尊重。天德贵人与月德贵人并见，称「双德入命」，一生福报深厚，德望高远。"},
		{"月德贵人", "ji", "月德贵人与天德贵人同为「德神」系列，以月支推算出阳干，若四柱中出现对应天干，即为命带月德。月德贵人主人心地善良，广结善缘，善于处理人际关系，在事业和社交中往往得到众多人的帮助与支持。此星能化解病灾、官讼、血光等凶险之事，具有强大的趋吉避凶功效。命带月德者，一生中多贵人相助，即便身处逆境，也能获得他人伸出援手。此星对女性命盘尤为有利，能化解产灾与家庭纷争，主夫妻和睦、子女孝顺。月德与天德并见，则福德双全，一生平安顺遂。"},
		{"天德合", "ji", "天德合是天德贵人所对应天干的六合干，同样具有类似天德贵人的化煞解厄功能，但力量稍弱于正神。命带天德合者，同样心性善良，处事圆融，能在逆境中得到帮助。此星具有合化的特性，主人善于协调周围关系，遇事能化解矛盾，将不利局面转化为机遇。在人际交往与社会公关方面有天赋，适合从事调解、谈判、外交、咨询等需要高情商的职业领域。天德合具有一定的护身作用，可减轻命中凶星的伤害，但需要命局整体格局配合方能充分发挥作用。"},
		{"月德合", "ji", "月德合是月德贵人对应天干的六合干，功能类似月德贵人，为化煞解灾的辅助吉神。命带月德合者，同样心性温厚，善结人缘，遇事多得助力。此星的合化属性使命主善于化解矛盾、融合各方关系，是天生的协调者与胶合剂。在团队合作、人际关系经营方面尤为出色，适合从事人力资源、公共关系、外交传播等领域。月德合虽力度略低于月德本身，但仍是积极的保护神，在日柱或时柱遇到凶煞时，能有效减轻伤害，使命盘整体运行趋于平稳与顺畅。"},
		{"德秀贵人", "ji", "德秀贵人是天德贵人与月德贵人的延伸神煞，凡命带天德或月德者，其所在柱同时得德秀贵人加持。德秀者，德行与秀气兼具之意，象征命主兼具内在美德与外在才华。此星主人仪表端庄，气质出众，同时心性善良，品德高洁，往往是「德才兼备」的典型。在社会交往中自带亲和力，自然吸引贵人青睐。命带德秀者，诚实守信，不走捷径，凭借真才实学赢得认可。此星在职场与婚姻中同样吉利，主配偶品质良好，上司提携，事业稳健上升。"},
		{"金舆贵人", "ji", "金舆贵人以日干帝旺之前一位推算，象征乘坐金辇（古代贵族专车）之意，主大富大贵、荣华富贵之命。命带金舆者，往往出身不凡或后天能跻身上层社会，享受物质生活的富足与精神生活的丰盛。此星主人仪态雍容，气度不凡，天生具有领导气质与贵族风度。在财富积累方面，金舆贵人预示着能获得意想不到的财富或显贵地位的人脉资源。男性命带金舆，主得贤妻助力；女性命带金舆，主嫁得好夫或自身贵显。此星宜与日柱或时柱同见，效力最强。"},
		{"天喜", "ji", "天喜星是专主喜悦、庆典与好消息的吉神，以年支推命，对应特定地支位置。命带天喜者，人生中喜事连连，婚喜、生育、升迁、金榜题名等喜庆之事较多。此星人缘极佳，社交活跃，擅长营造欢乐气氛，周围总是笑声不断。在感情方面，天喜主桃花旺盛，异性缘好，往往能吸引到优质的感情对象。流年逢天喜，是结婚、生子、乔迁、开业的大吉之年。女性命带天喜，主孕育顺遂，生育喜事多。天喜与红鸾（天赦）并见，感情与婚姻运势尤为兴旺，为命理「鸾喜双星」。"},
		{"天厨贵人", "ji", "天厨贵人以日干食神得禄之地推算，主衣食丰盛、口福极佳、一生不缺吃穿。古代命理认为，天厨入命者仿佛上天在命中为其设了一座永不断供的厨房。命带天厨者，口才极佳，擅长美食、烹饪与享受生活，往往在饮食、酒店、餐饮、娱乐行业有所成就。此星亦主精通手艺，多才多艺，有一技之长。在现代社会，天厨贵人者往往擅长厨艺、营养、食品或与口才相关的职业（如主播、演讲者）。一生温饱不愁，晚年生活尤为舒适惬意，是命理中代表「享受人生」的典型吉星。"},
		{"国印贵人", "ji", "国印贵人象征国家赋予的印章权力，代表官职、权威与社会地位的荣耀之星。以日干或各柱天干推算对应地支，出现在四柱即为命带国印。命带国印者，天生具有权威气场，适合从事政府机构、司法、管理、教育等需要权威认可的职业。此星主人得上司赏识，有机会获得职位晋升或重要任命，在行政与管理领域尤为得力。国印入日柱或时柱，主晚年位居要职或享有较高社会声望。此星不利于过于自由散漫的职业，更适合有组织体系、规则清晰的正规工作环境，是「公职命」的重要信号之一。"},
		{"三奇贵人", "ji", "三奇贵人分「天上三奇」（甲戊庚）与「地下三奇」（乙丙丁），若四柱天干中连续出现这些组合，即为三奇贵人。此星代表命中奇才异禀，往往有超凡的才智与创造力，是诸神煞中最为罕见的吉神之一。命带三奇者，思维超前，往往能在某一领域达到常人难以企及的高度，适合科研、发明、策划、哲学等需要创新思维的领域。三奇贵人者处事不落俗套，往往以独特视角解决问题，让人叹服。此星入命罕见，得此星者须有良好格局配合，方能将奇才转化为实际成就，否则可能流于「奇而无用」之虞。"},
		{"日德", "ji", "日德是指生于特定的日柱干支组合（甲寅、丙辰、戊辰、庚辰、壬戌等），代表日主自身携带的德行光华。命带日德者，天生具有高尚的人格品质，待人诚恳，处事公正，往往是所在群体中道德楷模式的人物。此星主人慈悲心重，乐于助人，积善积德，因此往往能得到来自各方的帮助与庇护。在职场上，日德者以诚信赢得信任，晋升路上少阴谋诡计，走的是堂堂正正的阳光大道。此星亦主健康长寿，因德行积累，晚年福报深厚。日德与天德、月德并见，则一生德望极高，声誉卓著。"},
		{"将星", "ji", "将星以年支三合局中的「帝旺」地支推算，若该支出现于四柱，即为命带将星。此星代表统帅、领导、掌权之象，是命理中主「权威与领导地位」的重要吉神。命带将星者，天生具有领导魅力，善于统筹调配，在团队中自然站到核心位置。此星适合从事军事、政治、管理、经营等需要掌控全局的职业。将星入命局有力者，往往在中年后获得实权，担任要职。古籍中将星代表将帅之命，自无全局统筹能力则容易被孤立。现代应用中，将星亦主在所在领域具有权威影响力，是事业成功的重要预兆。"},
		{"十灵日", "ji", "十灵日是命理中十个特殊干支日柱的专用神煞，包含甲子、甲午、甲申、甲戌、甲辰、甲寅六甲，以及乙亥、乙卯、乙丑、乙未四乙，共十组组合，故称「十灵」。生于十灵日者，被古代命理认为具有特殊的灵性与感知力，对超自然现象、玄学命理、宗教信仰等领域有天然的敏感与亲近感。此类人往往直觉敏锐，预感准确，能感知到常人无法察觉的信息。在心理学、玄学、医学（尤其中医）、哲学等探索生命本质的领域中，十灵日者往往有过人的见解与建树。此星不代表吉凶，而是一种特质的标志。"},
		{"词馆", "ji", "词馆贵人（又称词翰贵人）是主文采、学问与文职贵人的神煞，以日干推算对应地支，出现于四柱即为命带词馆。古代词馆指皇家翰林院，是文人学士得以施展才华的最高殿堂，因此词馆贵人象征命主具有文学才华、学术造诣，有机会凭借才学得到赏识与晋升。命带词馆者，文思敏捷，擅长写作、演讲、学术研究，在文化、教育、传媒、法律等文职领域容易脱颖而出。这些人往往能通过笔墨文章改变自身命运，在学界或文化界获得名声与地位。此星对学生、作家、律师、学者尤为有利。"},
		{"福星贵人", "ji", "福星贵人是专主福泽深厚、庇护人生的吉神，以年干推算对应地支，出现于四柱即为命带福星。与天乙贵人侧重「逢凶化吉」不同，福星贵人更侧重「平时增福」——让人的日常生活更加顺遂、机遇更多、福分更丰厚。命带福星者，往往家庭背景较好，从小得父母庇护，长大后人生道路相对平顺。此星亦主正财运旺盛，有稳定的收入来源，衣食不缺。在社会关系中，福星贵人者常被他人喜欢和帮助，无形中总有贵人助力。此星入日柱，主本命福泽深厚；入时柱，主晚年安享清福，子孙孝顺。"},
		{"天医", "ji", "天医星以月支推算，指特定月份的命主具有医疗缘分或医学天赋。命带天医者，往往对医学、中医、保健、心理等疗愈性领域有天然的亲近感与悟性。此星主人心细敏锐，善于观察他人身体与心理状态，是天生的医者或辅导者。在职业选择上，天医命主适合从事医疗、护理、心理咨询、中医养生、康复治疗等救助型工作。此星亦主命主本身健康运势较为良好，能及时察觉身体异常、防患未然。流年逢天医，是就医检查、调理身体的良好时机，也可能在医疗相关事务中取得重要进展或收获。"},
		{"羊刃", "xiong", "羊刃（又作「阳刃」）是日干在地支达到帝旺之地的神煞，象征锋芒毕露、力量过盛的状态。此星性质刚烈，带有强大的冲击力与主导欲，命带羊刃者往往性格强势，好胜心强，处事激进，容易与他人产生正面冲突。羊刃过旺则主克妻、克夫，伤害家庭关系，亦主外伤、手术、血光之灾。然而，羊刃并非全凶——七杀格遇羊刃，称「杀刃格」，反而刚劲有力，适合军警、外科、武术等需要刚毅之气的领域，有成就大事业的潜力。命局中羊刃需有制化，方能转危为安，化刚为用。无制则凶，有制则强悍可用。"},
		{"飞刃", "xiong", "飞刃是羊刃的对冲地支，象征外来的凶险冲击，是被动受害性质的神煞。与羊刃的「主动出击」不同，飞刃代表来自外部的碰撞与伤害，主意外事故、突发损伤、被人攻击或遭遇横祸。命带飞刃者，需特别注意交通安全、高空作业及各类意外风险，尤其在流年大运飞刃被冲触动之际。此星亦主感情世界多波折，与伴侣之间容易出现激烈冲突。飞刃落在日柱时，伤害最直接作用于日主本身，需特别警惕。格局好、命局稳健者，飞刃影响会相对减轻；格局弱者则需谨慎趋避，适时借助化解之道（如多行善积德、选择稳健的生活方向）。"},
		{"劫煞", "xiong", "劫煞是命理中主劫夺、损失与突发横祸的凶煞，以年支或日支的三合局起始位推算。命带劫煞者，人生中容易遭遇财物被劫、钱财损失、合同被毁、合伙纠纷等不愉快的劫夺经历。此星亦主暗中之敌，命主周围往往有嫉妒或暗算之人。在外出、投资、合伙等事项上需格外谨慎，避免被人欺骗或遭到意外损失。劫煞在年柱，童年或青少年时期家庭有动荡；在月柱，青壮年期事业多波折；在日柱，中年婚姻与健康受损；在时柱，晚年财物难保或子女不顺。有贵人星化解，可减轻劫煞的破坏力。"},
		{"亡神", "xiong", "亡神是代表消耗、散失与神志动荡的凶煞，以年支或日支推算。命带亡神者，精力容易过度消耗，事情往往功亏一篑，钱财难以积累，计划屡遭变故。此星亦主思想容易偏激，若八字整体格局不稳，则有过度追求玄学、宗教极端化或离群索居的倾向。亡神在感情方面主感情消散、分离、背叛，令人身心俱疲。然亡神并非完全无用，对于修行、研究、探索性工作（如哲学、密宗、侦探、心理分析）有独特的帮助，能使人深入内心世界，发现常人忽视的真相。需以格局整体评估，不可单论此星吉凶。"},
		{"孤辰", "xiong", "孤辰与寡宿是命理中象征孤独、离群的一对凶煞。孤辰以年支推算，主阳孤（男命孤单），寡宿主阴孤（女命孤寂）。命带孤辰者，内心深处有强烈的孤独感与疏离感，即便身处人群之中，也往往感到无人真正理解自己。孤辰者独立性强，自力更生能力极强，不依赖他人，但代价是在亲密关系上往往有所欠缺。此星在日柱时，主婚姻孤独感最强烈，夫妻易各过各的生活；在时柱，主晚年寂寥，子女缘分薄弱。从事出家、修行、哲学研究、独立创作等需要孤独沉淀的职业，孤辰反而能化为助力。"},
		{"寡宿", "xiong", "寡宿与孤辰相对，是命理中主孤独、寂寞的凶煞，尤对女命影响最深，古称「克夫之星」或「孀居之兆」。命带寡宿者，感情生活多有波折、孤苦，即便成婚也往往感情淡薄，或长期承受单身、丧偶、分居的孤寂。此星主人性格内敛含蓄，不善表达情感，难以建立深度亲密关系。然而从另一角度看，寡宿者往往心态沉稳，独立坚韧，适合专注于事业或学问的深耕。从事宗教信仰、学术研究、独立艺术创作等领域，寡宿反能成为心无旁骛、专一深入的助力。现代命理不宜完全以凶说论，需结合整体格局判断其影响程度。"},
		{"阴差阳错", "xiong", "阴差阳错是命理中极为特殊的神煞，主错位、误差与错综复杂的人生境遇。仅出现在少数特定的日柱干支组合中（如丙子、丁丑、戊寅、辛卯、壬辰、癸巳等），象征日主干支之间阴阳属性产生「差错」。命带阴差阳错者，人生中常有「差之毫厘、谬以千里」的经历——明明是正确的决策，却因细微误差导致截然不同的结果。在感情婚姻上阴差阳错最为明显，有缘无分、错过姻缘的情况较为常见，或婚姻中彼此理解有偏差。此星并非纯凶，属格局复杂之星，需整体分析命盘，方能准确判断其正负影响。"},
		{"魁罡", "xiong", "魁罡是命理中最为刚强霸气的神煞之一，仅出现于庚辰、庚戌、壬辰、壬戌四个日柱。此星代表极端刚毅、不妥协、掌控欲极强的性格特质，主聪明果断、文武双全，但脾气刚烈，不服管教，自我意志极为强烈。古籍称魁罡者「主聪明机辩，文武双全，但刚而过激」。男命魁罡，主事业心强，若得用，可成大器；女命魁罡，主性格刚强，婚姻感情多波折，传统命理认为「克夫」。现代来看，魁罡者适合军事、法律、执法、领导层等强势职业。命局中四柱见多个魁罡，刚性加倍，需特别注意情绪管理与人际关系协调。"},
		{"十恶大败", "xiong", "十恶大败是命理中最为忌讳的特殊日柱之一，包含甲辰、乙巳、丙申、丁亥、戊戌、己丑、庚辰、辛巳、壬申、癸亥共十组干支。命带此星者，被古书认为一生多逢不顺、挫折、消耗，事情常有始无终，计划功败垂成。此星亦主运势不稳定，逢财易散，遇官易祸，感情起伏不定。传统命理对十恶大败评价极为负面，然现代应用中发现，命盘整体格局良好、贵人神煞并见者，十恶大败的破坏力可大幅减轻。建议命带此星者，多积善行德，广结贵人，回避高风险的冒进投资与决策，以防运势被进一步削弱。"},
		{"天罗地网", "xiong", "天罗地网是命理中象征受困、缠缚与行动受阻的凶煞，以命盘中同时出现戌（天罗）与亥（地网）为判断标志。此星主人一生中容易陷入各种「网罗」之困——官司纠纷、感情纠缠、事业束缚、债务困境等，常有行动受限、出路被堵的压迫感。天罗地网在月柱最为明显，主中年事业多波折；在日柱，主婚姻感情多缠绕；在时柱，主老年处境被动。此星对男命影响大于女命，涉及法律诉讼时尤需谨慎。格局好、有化解神煞者，天罗地网的束缚力可减弱，但处事仍需格外谨慎，避免落入被动局面。"},
		{"地网", "xiong", "地网与天罗同为一对凶煞，以命盘中同时出现辰（地网）与巳（地网）为判断标志，主下层的困缚与牵绊。与天罗地网力量相近，地网主来自「地面世俗层面」的束缚，如财务陷阱、法律纠纷、人际关系困境等。命带地网者，容易被世俗的羁绊所困，如债务、契约、法律条款等，需特别谨慎处理所有书面合同与财务往来。此星落于哪一柱，该柱所代表的人生阶段（年柱-童年，月柱-青壮年，日柱-婚姻，时柱-晚年）相对容易遭遇困境。宜多行善积德，保持清廉正直，以化解地网的束缚之力。"},
		{"童子煞", "xiong", "童子煞是命理中象征婚姻延迟、感情障碍与孤克之性的神煞。命带童子煞者，古代认为有「不宜过早成婚」或「婚后感情淡薄」的倾向，部分命理认为此星与出家、修道或感情洁癖有关。日支在子、午、卯、酉（四旺地）者，或年支在子、午、卯、酉使时柱带煞者，均为童子入命。命带此星者，在感情上往往有完美主义倾向，对伴侣要求较高，难以满足，因此婚姻来得较晚，或婚后感情流于平淡。此星亦主人洁身自好，不喜沾染俗务，适合宗教、学术、艺术等清净领域。适当推迟婚期，可化解部分童子煞的影响。"},
		{"灾煞", "xiong", "灾煞是命理中主灾祸、突发事故与意外损伤的凶煞，以年支或日支的三合局来推算。命带灾煞者，人生中容易遭遇突如其来的灾难性事件，如意外事故、疾病突发、财产损失、手术开刀等。此星具有「无妄之灾」的特性，往往令人防不胜防。灾煞在年柱，主家族或童年有重大变故；在月柱，青壮年身心健康堪忧；在日柱，中年多身体方面的警示；在时柱，晚年或子孙有灾殃。命带灾煞者，宜坚持健康的生活方式，定期进行身体检查，从事高危职业时需加倍注意安全防护措施，日常生活中多留意交通与环境安全。"},
		{"流霞", "xiong", "流霞（又作霞光煞）以年干推算，主血光之灾、伤产与手术之厄的凶煞。古书称：「流霞主血光，女仅是产厄，男防刀伤」。命带流霞者，一生中与刀器、针划、手术、出血等事件缘分较深，需特别注意外伤与手术风险。对女命而言，流霞尤其与生育、妇科手术相关，生育时需特别留意。此星亦与消耗性疾病有关，如长期消耗精气神的病症。在流年大运中，流霞被触发时，是进行手术检查的高风险期，同时也可以借助「主动手术」（如计划性的手术）将被动's 血光转化为主动处理的契机，化解部分凶象。"},
		{"吊客", "xiong", "吊客是命理中主丧事、哀愁与生离死别的凶煞，以年支推算。命带吊客者，一生中与丧葬之事缘分不浅，容易经历亲人离世、朋友别离等令人悲伤的场景，也主情绪容易陷入低迷、忧郁状态。此星亦主贵人远离，助力减弱，在重要时刻较难得到充分的外部支持。吊客在年柱，主幼年家庭经历悲伤；在月柱，青壮年社交圈动荡；在日柱，中年情感创伤；时柱，晚年白发送黑发，子孙有离别之苦。流年逢吊客，是需要格外注意老人健康、谨慎出行的时期，不宜参与不吉祥的场合与庆典。"},
		{"墓门", "xiong", "墓门是命理中以天干五行克地支五行（天克地）之柱位推算的凶煞，象征该柱所代表生命阶段的压抑、受困与停滞。天克地意味着该柱干支之间存在五行相克关系，力量不能顺畅流通，产生内在紧张与阻碍。命带墓门者，所在柱位（年/月/日/时）代表的人生阶段容易遭遇较大的挑战与停滞，如事业受阻、健康警示、家庭纠纷等。此星亦主该柱所代表的六亲（如月柱对应父母兄弟，日柱对应配偶）关系较为紧张或有损耗。墓门并非极凶之星，结合格局整体判断，若命局强旺，此星的阻滞力可被化解；格局弱者需特别关注该生命阶段的健康与人际关系。"},
		{"桃花", "zhong", "桃花是命理中最为人熟知的情缘神煞，以年支或日支的三合局冠带之地推算，子午卯酉四支即为桃花星。命带桃花者，天生异性缘旺盛，社交魅力十足，面容姣好，性格迷人，容易吸引异性目光。桃花旺盛者，感情经历丰富，人生不缺爱情。在现代应用中，桃花不仅指感情，也主人际关系活跃、公众魅力强，适合从事演艺、公关、销售、服务业等需要良好亲和力的职业。然而桃花过旺或遇到凶星冲合，则可能流于「滥情」或「感情纠纷」。旺吉则是良缘，凶冲则宜节制。命盘整体格局决定桃花的表现方向。"},
		{"驿马", "zhong", "驿马是命理中主变动、奔波与旅行的神煞，以年支或日支的三合局起始之地推算。命带驿马者，天生喜欢变化，不甘安于一隅，人生中频繁变换居住地、工作岗位或生活环境。驿马旺者，往往从事需要出差、旅行、移居的职业，如外贸、旅游、运输、外交等。此星利于奔波谋利——越动越旺，越静则越不得力。驿马与贵人并见，则「贵人在远方」，异乡发展比本地更有利。反之，驿马与凶星并见，则奔波之中多损耗，折腾而无所获。现代命理将驿马与工作变动、移民出国挂钩，是判断人生流动性的重要参考指标。"},
		{"华盖", "zhong", "华盖原指古代皇帝出行时的伞盖，在命理中象征孤高、才华与神秘气质的神煞，以年支三合局的墓地推算。命带华盖者，天生具有独特的艺术气质与精神追求，往往对玄学、宗教、哲学、艺术等领域有浓厚兴趣，属于「曲高和寡」的人物。此星有孤独的一面——华盖者擅长独处，内心世界丰富，但与普通人沟通往往有隔阂，难以被完全理解。在命局中，华盖旺者适合从事宗教、艺术、学术、哲学等清高领域，有成为大师级人物的潜质。若命带华盖而不走孤高路线，则可能流于孤僻消沉。命局贵气强、华盖旺，则主超凡脱俗、名留青史。"},
		{"红艳", "zhong", "红艳（又称红鸾煞）是命理中主桃花异性缘与感情色彩的神煞，以日干推算对应地支。命带红艳者，外貌吸引力强，异性缘极旺，感情世界精彩丰富。此星兼具魅力与情欲双重属性，格局好者，红艳化为优雅的艺术气质与情感魅力；格局弱或凶星并见者，则红艳可能导致感情纠缠、桃色纠纷或情感上的执念过深。红艳与天乙贵人并见，感情中多得贵人相助；与羊刃并见，则感情锋芒过盛，易招惹口舌是非。在事业上，红艳赋予命主出色的艺术感与审美眼光，适合演艺、设计、时尚等彰显个人魅力的领域。命理中评价红艳需结合整体格局，不宜单凭此星论断感情命运。"},
	}

	for _, s := range shenshaSeedData {
		if _, err := DB.Exec(
			`INSERT INTO shensha_annotations (name, polarity, description) VALUES ($1, $2, $3) ON CONFLICT (name) DO NOTHING`,
			s.name, s.polarity, s.description,
		); err != nil {
			log.Printf("神煞注解 seed 失败 [%s]: %v", s.name, err)
		}
	}
	log.Println("✅ 神煞注解 (shensha_annotations) 初始化完成")

	// 增量迁移 (past-year-events)：过往年份事件推算报告缓存表
	pastEventsMigration := `
	CREATE TABLE IF NOT EXISTS ai_past_events (
		id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		chart_id       UUID NOT NULL UNIQUE REFERENCES bazi_charts(id) ON DELETE CASCADE,
		content_structured JSONB,
		model          VARCHAR(50),
		created_at     TIMESTAMPTZ DEFAULT NOW()
	);`
	if _, err := DB.Exec(pastEventsMigration); err != nil {
		log.Fatalf("增量迁移失败 (ai_past_events): %v", err)
	}

	// 增量迁移 (past-events-streaming-rewrite)：大运 summary 段独立缓存
	dayunSummaryMigration := `
	CREATE TABLE IF NOT EXISTS ai_dayun_summaries (
		id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		chart_id     UUID NOT NULL REFERENCES bazi_charts(id) ON DELETE CASCADE,
		dayun_index  INT NOT NULL,
		dayun_ganzhi VARCHAR(8),
		themes       JSONB,
		summary      TEXT,
		model        VARCHAR(50),
		created_at   TIMESTAMPTZ DEFAULT NOW(),
		UNIQUE (chart_id, dayun_index)
	);`
	if _, err := DB.Exec(dayunSummaryMigration); err != nil {
		log.Fatalf("增量迁移失败 (ai_dayun_summaries): %v", err)
	}

	log.Println("✅ 数据库迁移完成")
}


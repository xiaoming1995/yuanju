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
			description: "调候用神表（《穷通宝鉴》按月精华）",
			content: `调候用神表（《穷通宝鉴》按月精华节录）：

调候核心：先论寒暖燥湿，再定用神。命局过寒（冬月水旺）需丙丁火暖局；过热（夏月火旺）需壬癸水润局；过燥需水；过湿需火土。

甲木调候：
- 寅卯月（春）：木旺，取庚金劈甲引丁，丙癸辅助。
- 巳午月（夏）：木干，取癸水滋润为先，庚金次之。
- 申酉月（秋）：木衰，取丁火熟金裁木，庚丁同用。
- 亥子月（冬）：木冻，取丙火解冻为急，庚丁辅之。

丙火调候：
- 春月：壬水为用，有壬方显通明，配庚辛。
- 夏月：壬水为急，防炎上格，取壬济之。
- 秋月：壬水、甲木为用（秋火弱需木生）。
- 冬月：甲木引丁，壬水辅助，防寒极无根。

壬水调候：
- 春月：取戊土为堤防，庚辛为源头。
- 夏月：取庚辛为源，再取甲木疏土透水。
- 秋月：甲木疏泄，戊土制水，避免泛滥。
- 冬月：戊土制水为要，内忌比劫争合。

断法要点：调候用神若透干得地，命主体质健旺、运途顺达；若调候用神被冲克，则体弱多病、运势多阻。`,
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
	}

	for _, seed := range kbSeeds {
		if _, err := DB.Exec(
			`INSERT INTO ai_prompts (module, content, description) VALUES ($1, $2, $3) ON CONFLICT (module) DO NOTHING`,
			seed.module, seed.content, seed.description,
		); err != nil {
			log.Printf("初始化知识模块 [%s] 失败: %v", seed.module, err)
		}
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

	log.Println("✅ 数据库迁移完成")
}

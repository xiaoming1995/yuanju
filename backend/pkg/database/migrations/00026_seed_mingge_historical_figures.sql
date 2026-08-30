-- +goose Up

-- These are editorial Ming Ge references, not claims that a reader matches a
-- historical figure. Birth-time data is intentionally marked unknown, so no
-- Dayun assertion is published for the initial collection.
INSERT INTO mingge_historical_figures (
	ming_ge, figure_name, era, identity, historical_memory, turning_point, turning_point_year,
	source_title, source_url, birth_data_precision, bazi_verification_note,
	show_dayun, display_order, review_status
) VALUES

('伤官格', '李白', '唐代', '诗人', '以大量传世诗作、强烈的想象力和自由奔放的表达，被后世视为唐诗的重要代表。', '受召入翰林，作品与声名进入更广阔的公共视野。', '742 年前后', 'Poetry Foundation：Li Po', 'https://www.poetryfoundation.org/poets/li-po', 'unknown', '未按出生时辰核验；作为表达与创造方向的古人映照，不展示大运。', false, 10, 'published'),
('伤官格', '苏轼', '北宋', '文学家、官员', '以诗、词、散文、书画及豁达的生命态度被后世记住，是宋代文化中的重要人物。', '进士及第后进入仕途，逐步形成兼具政务与文学创作的长期影响。', '1057 年起', 'Encyclopaedia Britannica：Su Shi', 'https://www.britannica.com/biography/Su-Shi', 'unknown', '未按出生时辰核验；作为表达与创造方向的古人映照，不展示大运。', false, 20, 'published'),
('食神格', '陶渊明', '东晋', '诗人', '以田园诗和归隐书写形成独特的审美传统，作品长期影响中国文学。', '辞去官职后回归乡里，以持续写作确立田园诗人的文化形象。', '405 年前后', 'Encyclopaedia Britannica：Tao Yuanming', 'https://www.britannica.com/biography/Tao-Yuanming', 'unknown', '未按出生时辰核验；作为稳定输出与生活审美方向的古人映照，不展示大运。', false, 10, 'published'),
('食神格', '白居易', '唐代', '诗人、官员', '以语言平易而关怀现实的诗作广为流传，其作品在东亚文化圈长期被传诵。', '入仕与持续写作并行，形成兼具公共关怀与文学影响力的创作道路。', '贞元末至元和年间', 'Encyclopaedia Britannica：Bai Juyi', 'https://www.britannica.com/biography/Bai-Juyi', 'unknown', '未按出生时辰核验；作为稳定输出与生活审美方向的古人映照，不展示大运。', false, 20, 'published'),
('正财格', '管仲', '春秋', '政治家、改革者', '辅佐齐桓公整顿政务与经济，推动齐国成为春秋时期的重要诸侯国。', '受鲍叔牙举荐后辅佐齐桓公，主持内政与经济制度建设。', '前 685 年后', 'Encyclopaedia Britannica：Guan Zhong', 'https://www.britannica.com/biography/Guan-Zhong', 'unknown', '未按出生时辰核验；作为务实经营与资源组织方向的古人映照，不展示大运。', false, 10, 'published'),
('正财格', '张居正', '明代', '政治家、改革者', '以整饬吏治、财政改革和推行考成法，被视为明代后期重要的改革人物。', '出任内阁首辅后推进考成法与财政整顿，改革影响扩大。', '1572 年后', 'Encyclopaedia Britannica：Zhang Juzheng', 'https://www.britannica.com/biography/Zhang-Juzheng', 'unknown', '未按出生时辰核验；作为务实经营与资源组织方向的古人映照，不展示大运。', false, 20, 'published'),
('偏财格', '吕不韦', '战国末期', '商人、政治人物', '从经商走向政治运作，在秦国早期权力格局中留下显著影响。', '以资源整合与人际网络进入秦国政治核心，主持编纂《吕氏春秋》。', '前 3 世纪中后期', 'Encyclopaedia Britannica：Lu Buwei', 'https://www.britannica.com/biography/Lu-Buwei', 'unknown', '未按出生时辰核验；作为机会资源与人际组织方向的古人映照，不展示大运。', false, 10, 'published'),
('偏财格', '郑和', '明代', '航海家、外交使者', '多次远航连接亚非多地，成为中国航海史与跨区域交流中的代表人物。', '受命率领船队下西洋，承担大规模航海、外交与物资组织任务。', '1405 年起', 'Encyclopaedia Britannica：Zheng He', 'https://www.britannica.com/biography/Zheng-He', 'unknown', '未按出生时辰核验；作为机会资源与跨区域组织方向的古人映照，不展示大运。', false, 20, 'published'),
('正官格', '狄仁杰', '唐代', '官员、政治家', '以断案与辅政声名流传，常被后世作为秩序、责任与识人能力的历史形象。', '在武周时期历任要职并参与政务决策，影响力显著提升。', '7 世纪末', 'Encyclopaedia Britannica：Di Renjie', 'https://www.britannica.com/biography/Di-Renjie', 'unknown', '未按出生时辰核验；作为秩序、责任与制度执行方向的古人映照，不展示大运。', false, 10, 'published'),
('正官格', '包拯', '北宋', '官员', '因清廉、公正的形象深入民间叙事，成为司法公信与秩序意识的文化符号。', '任开封府尹及御史等职时，以敢于直言和执法形象为后世所记。', '11 世纪中叶', 'Encyclopaedia Britannica：Bao Zheng', 'https://www.britannica.com/biography/Bao-Zheng', 'unknown', '未按出生时辰核验；作为秩序、责任与制度执行方向的古人映照，不展示大运。', false, 20, 'published'),
('七杀格', '岳飞', '南宋', '军事统帅', '以抗金军事实践和忠勇形象被后世长期纪念，是中国历史叙事中的重要人物。', '组织岳家军并在南宋初年多次参与北伐与防御作战。', '1130 年代', 'Encyclopaedia Britannica：Yue Fei', 'https://www.britannica.com/biography/Yue-Fei', 'unknown', '未按出生时辰核验；作为压力下行动、竞争与担当方向的古人映照，不展示大运。', false, 10, 'published'),
('七杀格', '戚继光', '明代', '军事家', '组织练兵并抗击倭寇，其军事训练与著述在后世持续被讨论。', '在东南沿海整训军队、参与抗倭，形成较成熟的练兵体系。', '16 世纪中叶', 'Encyclopaedia Britannica：Qi Jiguang', 'https://www.britannica.com/biography/Qi-Jiguang', 'unknown', '未按出生时辰核验；作为压力下行动、竞争与担当方向的古人映照，不展示大运。', false, 20, 'published'),
('正印格', '朱熹', '南宋', '思想家、教育家', '系统阐释理学并长期从事讲学与著述，对后世教育与思想史影响深远。', '长期讲学、整理经典并形成有系统的理学论述，学术影响逐渐扩大。', '12 世纪后期', 'Encyclopaedia Britannica：Zhu Xi', 'https://www.britannica.com/biography/Zhu-Xi', 'unknown', '未按出生时辰核验；作为学习、传承与正统资源方向的古人映照，不展示大运。', false, 10, 'published'),
('正印格', '蔡伦', '东汉', '官员、技术改进者', '以改进造纸工艺的传统记载被后世广泛认知，成为技术传播史的重要人物。', '在宫廷任职期间主持或参与改进造纸工艺，使相关技术更易推广。', '2 世纪初', 'Encyclopaedia Britannica：Cai Lun', 'https://www.britannica.com/biography/Cai-Lun', 'unknown', '未按出生时辰核验；作为学习、传承与正统资源方向的古人映照，不展示大运。', false, 20, 'published'),
('偏印格', '张衡', '东汉', '科学家、文学家、官员', '在天文、数学、机械与文学领域均有贡献，是中国古代跨学科探索的代表人物。', '在太史令等职位上参与天文观测与仪器研究，形成多领域成果。', '2 世纪前期', 'Encyclopaedia Britannica：Zhang Heng', 'https://www.britannica.com/biography/Zhang-Heng', 'unknown', '未按出生时辰核验；作为研究、独特思路与非标准路径方向的古人映照，不展示大运。', false, 10, 'published'),
('偏印格', '徐霞客', '明代', '旅行家、地理学者', '长期实地考察山川地貌，留下《徐霞客游记》，成为地理考察与游记写作的代表。', '长期远行实察并整理行旅笔记，形成以实地观察为核心的著述。', '17 世纪前期', 'Encyclopaedia Britannica：Xu Xiake', 'https://www.britannica.com/biography/Xu-Xiake', 'unknown', '未按出生时辰核验；作为研究、独特思路与非标准路径方向的古人映照，不展示大运。', false, 20, 'published'),
('建禄格', '王阳明', '明代', '思想家、官员、军事家', '以心学思想、地方治理与军事行动并存的经历，被视为知行合一的代表人物。', '平定地方动乱并讲学传播心学，形成思想与实践并行的影响。', '16 世纪初', 'Encyclopaedia Britannica：Wang Yangming', 'https://www.britannica.com/biography/Wang-Yangming', 'unknown', '未按出生时辰核验；作为自驱、持续推进与独立实践方向的古人映照，不展示大运。', false, 10, 'published'),
('建禄格', '曾国藩', '清代', '官员、军事统帅', '在晚清政局与地方军事组织中发挥重要作用，其修身与组织实践被后世反复讨论。', '组织湘军并在晚清政局中承担更大责任，形成长期组织影响。', '19 世纪中期', 'Encyclopaedia Britannica：Zeng Guofan', 'https://www.britannica.com/biography/Zeng-Guofan', 'unknown', '未按出生时辰核验；作为自驱、持续推进与独立实践方向的古人映照，不展示大运。', false, 20, 'published'),
('月刃格', '霍去病', '西汉', '军事统帅', '以年轻统帅身份参与对匈奴作战并取得重要战功，成为汉代军事史中的鲜明人物。', '率军多次远征并取得河西等战役成果，军事声望迅速建立。', '前 2 世纪', 'Encyclopaedia Britannica：Huo Qubing', 'https://www.britannica.com/biography/Huo-Qubing', 'unknown', '未按出生时辰核验；作为强行动力与风险驾驭方向的古人映照，不展示大运。', false, 10, 'published'),
('月刃格', '辛弃疾', '南宋', '词人、军事人物', '以豪放词和抗金抱负闻名，作品中长期保留强烈的行动意志与家国主题。', '参与抗金行动后南归，并以词作与政务经历形成持久文化影响。', '12 世纪后期', 'Encyclopaedia Britannica：Xin Qiji', 'https://www.britannica.com/biography/Xin-Qiji', 'unknown', '未按出生时辰核验；作为强行动力与风险驾驭方向的古人映照，不展示大运。', false, 20, 'published')
ON CONFLICT (ming_ge, figure_name) DO NOTHING;

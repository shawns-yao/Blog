/**
 * 前台主导航（硬编码）。
 *
 * 「导航菜单」整块已删除：不再有 nav_menu 表与 /public/nav-menus 接口，
 * 入口列表固定在代码内，数组顺序即渲染顺序。
 *
 * icon 取值必须是 $lib/ui/icons/lucide-loaders 白名单中的键，
 * 否则 DynamicLucideIcon 会渲染为空。
 */

export interface NavItem {
	name: string;
	url: string;
	icon?: string | null;
	children?: NavItem[];
}

export const NAV_ITEMS: NavItem[] = [
	{ name: '首页', url: '/', icon: 'house' },
	{ name: '手记', url: '/moments', icon: 'pen-tool' },
	{ name: '图书馆', url: '/gallery', icon: 'book-open' },
	{ name: '音乐室', url: '/music', icon: 'music' },
	{ name: '友链', url: '/friends', icon: 'link' },
	{ name: '关于', url: '/about', icon: 'user' }
];

/**
 * 书架导航的一本书。
 *
 * 视觉隐喻：书立在木质层板上，静止时各自带一点倾斜角度（东倒西歪），
 * 悬停时扶正并微微抬起。角度与高度刻意做成不整齐，避免「罗列感」。
 */
export interface ShelfBook {
	name: string;
	url: string;
	/** 必须是 lucide-loaders 白名单中的键 */
	icon: string;
	/** 书脊底色 */
	color: string;
	/** 书脊顶部深色描边 */
	edge: string;
	/** 静止时的倾斜角度（deg），正负交替 */
	tilt: number;
	/** 书高（px），略有差异更自然 */
	height: number;
	/** 书脊宽度（px） */
	width: number;
	/** 书在层板上的摆放方式 */
	placement?: 'upright' | 'flat';
	/** 平铺书相对层板抬高的距离（px），用于叠放 */
	lift?: number;
}

/** 书架上的顶级书。数组顺序与摆放方式共同决定构图。 */
export const SHELF_BOOKS: ShelfBook[] = [
	{
		name: '首页',
		url: '/',
		icon: 'house',
		color: '#874a34',
		edge: '#5d2f23',
		tilt: 6,
		height: 132,
		width: 52
	},
	{
		name: '手记',
		url: '/moments',
		icon: 'feather',
		color: '#526551',
		edge: '#344337',
		tilt: 4,
		height: 120,
		width: 45
	},
	{
		name: '图书馆',
		url: '/gallery',
		icon: 'book-open',
		color: '#3f5871',
		edge: '#293c51',
		tilt: 7,
		height: 127,
		width: 48
	},
	{
		name: '音乐室',
		url: '/music',
		icon: 'music',
		color: '#9a7442',
		edge: '#6c4f2e',
		tilt: 10,
		height: 130,
		width: 51
	},
	{
		name: '友链',
		url: '/friends',
		icon: 'link',
		color: '#725144',
		edge: '#4b3029',
		tilt: -1,
		height: 31,
		width: 94,
		placement: 'flat',
		lift: 0
	},
	{
		name: '关于',
		url: '/about',
		icon: 'user',
		color: '#655263',
		edge: '#453747',
		tilt: 2,
		height: 29,
		width: 90,
		placement: 'flat',
		lift: 29
	}
];

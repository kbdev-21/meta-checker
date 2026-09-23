// Icon position của Riot. Vite hash tên file lúc build nên import thay vì để trong static/.
import adc from './assets/positions/adc.png';
import jgl from './assets/positions/jgl.png';
import mid from './assets/positions/mid.png';
import spt from './assets/positions/spt.png';
import top from './assets/positions/top.png';

import type { Position } from '$lib/api';

// UNK (ARAM, Arena...) không có icon; UI tự fallback về chữ.
export const POSITION_ICONS: Record<Position, string | null> = {
	TOP: top,
	JGL: jgl,
	MID: mid,
	ADC: adc,
	SPT: spt,
	UNK: null
};

// Icon position của Riot. Vite hash tên file lúc build nên import thay vì để trong static/.
import adc from '$lib/assets/positions/adc.png';
import jgl from '$lib/assets/positions/jgl.png';
import mid from '$lib/assets/positions/mid.png';
import spt from '$lib/assets/positions/spt.png';
import top from '$lib/assets/positions/top.png';

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

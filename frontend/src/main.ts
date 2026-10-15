import './style.css';
import {subscribe} from './store';
import {render} from './render';

subscribe(render);

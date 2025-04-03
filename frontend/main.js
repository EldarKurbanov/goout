import Calendar from '@event-calendar/core';
import TimeGrid from '@event-calendar/time-grid';
import '@event-calendar/core/index.css';

const ec = new Calendar({
    target: document.getElementById('ec'),
    props: {
        plugins: [TimeGrid],
        options: {
            view: 'timeGridWeek',
            events: [
                { title: 'Встреча', start: '2025-04-10T10:00:00', end: '2025-04-10T11:30:00' },
                { title: 'Звонок', start: '2025-04-12T14:00:00', end: '2025-04-12T15:30:00' }
            ]
        }
    }
});

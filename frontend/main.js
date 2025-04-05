import Calendar from '@event-calendar/core';
import TimeGrid from '@event-calendar/time-grid';
import '@event-calendar/core/index.css';

const ec = new Calendar({
    target: document.getElementById('ec'),
    props: {
        plugins: [TimeGrid],
        options: {
            view: 'timeGridWeek',
            eventSources: [
                {
                    url: 'http://localhost/events',
                    method: 'GET',
                }
            ]
        }
    }
});

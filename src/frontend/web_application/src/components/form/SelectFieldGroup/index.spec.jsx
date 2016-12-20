import React from 'react';
import { shallow } from 'enzyme';
import SelectFieldGroup from './';

describe('component SelectFieldGroup', () => {
  it('render', () => {
    const props = {
      label: 'Foo',
      options: [{ label: 'a', value: 1 }, { label: 'b', value: 3 }],
      onChange: jasmine.createSpy('onChange'),
    };

    const comp = shallow(
      <SelectFieldGroup {...props} />
    );

    expect(comp.find('label').text()).toEqual('Foo');
    expect(comp.find('option').length).toEqual(2);
    comp.find('select').simulate('change');
    expect(props.onChange).toHaveBeenCalled();
  });
});
